
```

oc login -u system:admin
make cluster/clean
make cluster/prepare

make code/run


TODO:
oc create clusterrole mobileclient-admin --verb=create,delete,get,list,patch,update,watch --resource=mobileclients
oc adm policy add-cluster-role-to-user mobileclient-admin system:serviceaccount:mdc:example-mobiledeveloperconsole 

kubectl apply  -n mdc -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_cr.yaml
```