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
const axios = require('axios');
const AxiosLogger = require('axios-logger');

const { PubSub } = require('@google-cloud/pubsub');
const pubSubClient = new PubSub();
topicName = 'quote-data',

  exports.fetchTickers = (req, res) => {
    var date = req.query.hasOwnProperty('date') ? req.query.date : undefined;
    __fetchTickersV2(date).then(function (quotes) {
      const dataBuffer = Buffer.from(quotes);

      async function publishMessage(data) {
        // Publishes the message as a string, e.g. "Hello, world!" or JSON.stringify(someObject)
        const dataBuffer = Buffer.from(data);

        try {
          const messageId = await pubSubClient.topic(topicName).publish(dataBuffer);
          console.log(`Message ${messageId} published.`);
        } catch (error) {
          console.error(`Received error while publishing: ${error.message}`);
          process.exitCode = 1;
        }
      }

      publishMessage(quotes);

      res.status(200).send(quotes);
    }, function (err) {
      res.status(400).send(err);
    });
  }
var errHandler = function (err) {
  console.log(err);
}

process.on('unhandledRejection', err => {
  console.error(err.message);
  process.exitCode = 1;
});

function __fetchTickers(date) {
  if (date) {
    date = new Date(date);
    console.log("Received Request Date " + date);
  } else {
    date = new Date();//  new Date('1995-12-17')
  }
  // The URL Should be of format http://www.nseindia.com/content/historical/EQUITIES/2019/JUN/cm07JUN2019bhav.csv.zip'
  var file_url = 'https://www1.nseindia.com/content/historical/EQUITIES/' + date.getFullYear() + '/' + months[date.getMonth()] + '/cm' + getDate(date) + 'bhav.csv.zip';
  var dataPromise = getData(file_url);
  dataPromise.then(function (quotes) {
    writeToBucket(date, quotes);
    return quotes;
  }, errHandler);
  return dataPromise;
}

function __fetchTickersV2(date) {
  if (date) {
    date = new Date(date);
    console.log("Received Request Date " + date);
  } else {
    date = new Date();//  new Date('1995-12-17')
  }

  // Sample file url 
  // https://nsearchives.nseindia.com/products/content/sec_bhavdata_full_14072024.csv
  const month =
    (date.getMonth() + 1).toLocaleString('en-US',
      { minimumIntegerDigits: 2, useGrouping: false });
  const day =
    date.getDate().toLocaleString('en-US',
      { minimumIntegerDigits: 2, useGrouping: false });

  var file_url = 'https://nsearchives.nseindia.com/products/content/sec_bhavdata_full_'
    + day + month + date.getFullYear() + '.csv';
  var dataPromise = getCsvData(file_url);
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

    var headers = {
      'Referer': 'https://www1.nseindia.com/products/content/equities/equities/archieve_eq.htm',
      'user-agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36'
    };

    request.get({ url: file_url, encoding: null, headers: headers }, (err, res1, body) => {
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

function getCsvData(file_url) {
  console.log("Downloading from - axios " + file_url);

  axios.interceptors.request.use( AxiosLogger.requestLogger);
  AxiosLogger.setGlobalConfig({
    prefixText: 'NASE',
    dateFormat: 'HH:MM:ss',
    status: true,
    headers: true,
    data : false,
});
  axios.interceptors.response.use(AxiosLogger.responseLogger, AxiosLogger.errorLogger);

  return new Promise(function (resolve, reject) {

    axios.get(file_url, {
      timeout: 1000 * 5
      , headers: {
        'Referer': 'https://www.nseindia.com/all-reports',
        'user-agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36'
      }
    })
      .then(function (response) {
        if (response.statusCode >= 400) {
          err = "Did not find any price for " + file_url + response;
          reject(err);
        } else {
          console.log("Found prices size " + response.data.length);
          resolve(response.data);
        }
      })
      .catch(function (error) {
        err = "Did not find any price for " + file_url + error;
        reject(err);
      })
      .finally(function () {
        console.log("Download operation completed OK/Error");
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
  console.log(start + " & " + end);
  while (loop <= end) {
    console.log("Fetching data for " + loop);
    __fetchTickers(loop.toISOString()).then(function (quotes) {
      console.log(quotes.length);
    }, (err) => { console.log("Error Found " + err) });
    var newDate = loop.setDate(loop.getDate() + 1);
    loop = new Date(newDate);
  }
}

