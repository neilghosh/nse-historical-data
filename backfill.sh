#!/bin/bash
# This does not work in Mac
d="2021-01-01"
until [[ $d > 2021-04-17 ]]; do 
    echo "$d"
    url='https://us-central1-demoneil.cloudfunctions.net/getQuotesByDate?date='$d
    echo $url
    curl -s ${url} --output /dev/null 
    d=$(date -I -d "$d + 1 day")
done
