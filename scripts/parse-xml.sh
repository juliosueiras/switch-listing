cat switchtdb.xml | xq --xml-force-list locale -r '.datafile.game[] | "\(.locale[] | select(.["@lang"] == "EN") | .title)|\(.id)|\(.region)"' 
