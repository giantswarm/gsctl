# coding: utf8

# This script helps to fetch a list of AWS EC2 instance types.

import requests

url = "https://raw.githubusercontent.com/powdahound/ec2instances.info/master/www/instances.json"

r = requests.get(url)
r.raise_for_status()

# generate Go map
print('\tAwsEc2InstanceTypes = map[string]bool{')
for item in r.json():
    print('\t\t"%s": true,' % item["instance_type"])
print('\t}')
