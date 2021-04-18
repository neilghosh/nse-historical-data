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

## Deployed as cloud function 

Dowload Backfill
```
d="2020-01-04"
until [[ $d > 2020-07-23 ]]; do 
    echo "$d"
    url='https://us-central1-demoneil.cloudfunctions.net/getQuotesByDate?date='$d
    echo $url
    curl -s ${url} --output /dev/null 
    d=$(date -I -d "$d + 1 day")
done

```

See dailt data count 

```
SELECT count(*), TIMESTAMP FROM `demoneil.nse_data.nse_historical_data` group by TIMESTAMP order by TIMESTAMP DESC LIMIT 1000
```

## Create View of Data for unique date
```
     SELECT *
     FROM (
           SELECT
           *,
               ROW_NUMBER()
                   OVER (PARTITION BY SYMBOL, TIMESTAMP)
                   row_number
           FROM demoneil.nse_data.nse_historical_data
           WHERE SERIES = 'EQ'
         )
     WHERE row_number = 1
     -- SELECT  FROM `demoneil.nse_data. nse_historical_data_unique` LIMIT 1000
```


## Delete duplicate data
```
CREATE OR REPLACE TABLE `demoneil.nse_data.nse_historical_data_dedupe`
AS
SELECT * FROM `demoneil.nse_data.nse_historical_data`
```

## saveQuoteToDatastore
Automatically saves quote to DS
![image](https://user-images.githubusercontent.com/726337/115161330-ba9acb00-a0ba-11eb-91b9-79c708b699cd.png)
