/**
 * Responds to any HTTP request.
 *
 * @param {!express:Request} req HTTP request context.
 * @param {!express:Response} res HTTP response context.
 */
const months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
const { Storage } = require('@google-cloud/storage');
const storage = new Storage();
const bucketName = 'nse-historical-prices';
var gcsBucket = storage.bucket(bucketName);

exports.fetchTickers = (req, res) => {
  var date = req.query.hasOwnProperty('date') ? req.query.date : undefined;
  __fetchTickers(date).then(function (quotes) {
    res.status(200).send(quotes);
  }, function (err) {
    res.status(400).send(err);
  });
}
var errHandler = function (err) {
  console.log(err);
}

function __fetchTickers(date) {
  if (date) {
    date = new Date(date);
    console.log("Received Request Date " + date);
  } else {
    date = new Date();//  new Date('1995-12-17')
  }
  // The URL Should be of format http://www.nseindia.com/content/historical/EQUITIES/2019/JUN/cm07JUN2019bhav.csv.zip'
  var file_url = 'http://www1.nseindia.com/content/historical/EQUITIES/' + date.getFullYear() + '/' + months[date.getMonth()] + '/cm' + getDate(date) + 'bhav.csv.zip';
  var dataPromise = getData(file_url);
  dataPromise.then(function (quotes) {
    writeToBucket(date, quotes);
    return quotes;
  }, errHandler);
  return dataPromise;
}

function getData(file_url) {
  console.log("Downloading from " + file_url);
  var AdmZip = require('adm-zip');
  var request = require('request');

  // Return new promise 
  return new Promise(function (resolve, reject) {
    // Do async job
    request.get({ url: file_url, encoding: null }, (err, res1, body) => {
      if (err) {
        reject(err);
      } else if (res1.statusCode >= 400) {
        err = "Did not find any price for " + file_url;
        reject(err);
      } else {
        console.log("Found teh archived prices size " + body.length);
        var zip = new AdmZip(body);
        var zipEntries = zip.getEntries();
        console.log("Found content of length " + zipEntries.length);
        entry = zipEntries[0];
        var quotes = zip.readAsText(entry);
        resolve(quotes);
      }
    });
  })
}

function getDate(date) {
  return ("0" + date.getDate()).slice(-2) + months[date.getMonth()] + date.getFullYear();
}

async function writeToBucket(date, data) {
  var fileName = date.toISOString().split('T')[0] + '.csv';
  console.log("Saving in GCS as " + fileName);
  var file = gcsBucket.file(fileName);

  file.createWriteStream({
    metadata: {
      contentType: 'text/csv',
      metadata: {
        custom: 'metadata'
      }
    }
  })
    .on('error', function (err) { })
    .on('finish', function () { console.log("Uploaded"); })
    .end(data);
}

function getDataInRange(startDate, endDate) {
  var start = new Date(startDate);
  var end = new Date(endDate);
  var loop = new Date(start);
  console.log(start+" & "+end);
  while (loop <= end) {
    console.log("Fetching data for "+loop);
    __fetchTickers(loop.toISOString()).then(function (quotes) {
      console.log(quotes.length);
    }, (err) => { console.log("Error Found " + err) });
    var newDate = loop.setDate(loop.getDate() + 1);
    loop = new Date(newDate);
  }
}

