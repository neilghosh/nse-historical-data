# nse-historical-data

NodeJS code can be deployed as part of cloud function to download NSE historical data into a GCS bucket . CSV files with data can be found in the bucket with file name YYYY-MM-DD.csv

## Setup

1. Create a GSC Bucket and set the name in the code.

## Run Locally 

###
Install Node 
``` brew install node ```
Install Dependent Packages
``` npm install ```

### Set default credientials of GCP Project 
```
gcloud auth application-default login
```
### Invoke with argument 

Download Data for Today 
```
node index.js 
```

Download Data for Past Date
```
node index.js 2019-08-01
```

Download Data for Past Date Range
```
node index.js 2019-08-03
```

## Experimental 
Run from command line without helper function 
```
node -e 'require("./index").fetchTickers({query:{date:"2019-07-31"}},{status:() => ({})})'
```
