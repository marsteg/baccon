ID="0aeb735c-5327-4c35-a5ef-271d66338657"
while true
 do 
    curl https://baccon.apps.cp-ccee.trinity.shoot.canary.k8s-hana.ondemand.com/postgres/$ID/write
    sleep 0.1
    curl https://baccon.apps.cp-ccee.trinity.shoot.canary.k8s-hana.ondemand.com/postgres/$ID/query -s >/dev/null
    sleep 0.1
done