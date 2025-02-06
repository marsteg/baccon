#!/bin/bash



result=$(curl -s -X POST https://baccon.apps.cp-ccee.trinity.shoot.canary.k8s-hana.ondemand.com/postgres/ -d '{
  "host": "10.180.55.88",
  "port": "5432",
  "name": "wurst9",
  "user": "user",
  "password": "password",
  "database": "postgres"
}')

ID=$(echo $result | jq .id | tr -d '"')

while true
 do 
    curl https://baccon.apps.cp-ccee.trinity.shoot.canary.k8s-hana.ondemand.com/postgres/$ID/write
    sleep 0.1
    curl https://baccon.apps.cp-ccee.trinity.shoot.canary.k8s-hana.ondemand.com/postgres/$ID/query -s >/dev/null
    sleep 0.1
done
