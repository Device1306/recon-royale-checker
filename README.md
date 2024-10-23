a simple script to help check results for the recon-royale.com challenge.

```
# I recommend only using files with unique targets make sure to do something like cat combine_list.txt | sort -u > unique_list.txt
# run the checker on the target input file (this should be all the identified targets from other recon tools)
go run checker.go -f inputfile.txt

# Using wc -l you can check the total count of items in the file
go run checker.go -f inputfile.txt | wc -l

# Run the checker and output the items to a file
go run checker.go -f inputfile.txt -o outputfile.txt

# cat the file output in wc to see word count.
cat outputfile.txt | wc -l
```
