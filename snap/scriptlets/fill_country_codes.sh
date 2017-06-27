#!/bin/sh
# The purpose of this script is taking ISO-3166 country codes from country_code file
# so that they can be updated in management portal when snap is built
# Process is easy, get it from remote path, format using sed, and replace 
# html page were they will be used

set -e 

COUNTRY_CODES=../../../snap/scriptlets/country-codes

if [ ! -e $COUNTRY_CODES ]; then
    echo "======================================================="
    echo "Could not find country_codes needed file for compiling."
    echo "Please, before building snap for the first time execute by hand:"
    echo ""
    echo "snap/scriptlets/fetch_country_codes.sh"
    echo ""
    echo "======================================================="
    exit 1
fi

# think that this is a scriptlet, executed in parts/<the_part>/build folder
cp $COUNTRY_CODES .

# remove non processable lines
sed -i '/^<a href=iso/!d' country-codes

# process lines to change its format to be html select options
sed -i '/<a href=iso/ s/<a href=iso.*html>/<option value="/ 
s/<\/a>\ :\ /">/
s/<br>/<\/option>/' country-codes

# add world wide default option at beginning
sed -i '1s;^;<option value="XX">\-WORLD WIDE\-<\/option>\n;' country-codes

# replace mark in management.html template with real country-codes
sed -e '/\[COUNTRY_CODE_OPTIONS\]/ {' -e 'r country-codes' -e 'd' -e '}' -i static/templates/management.html 

echo "country codes filled ok"
