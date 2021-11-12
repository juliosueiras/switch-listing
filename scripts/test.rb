require 'csv'

CSV::open("result2.csv", :headers => true).each {|i| 
  `xmlstarlet edit -L -s "/collectorz/data/gameinfo/gamelist/game/index[text()='#{i["Index"]}']/.." -t elem -n notes -v '#{i["ProductCode"]}' test.xml`
}

