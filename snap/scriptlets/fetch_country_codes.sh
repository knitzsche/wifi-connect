#!/bin/sh
# The purpose of this script is taking ISO-3166 country codes and
# store them locally to be used in compilation time.
# This operation must be executed by hand when wanted to update
# current ones shown in config page
set -e 

cd "$(dirname "$0")"

curl -X GET http://geotags.com/iso3166/countries.html > country-codes