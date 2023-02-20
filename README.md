# td-visa-pdf-xactions-to-csv

Extract transactions from TD Visa statement

This project came about because every year I need to get all my credit card transactions for the
last income tax and categorize them for tax credits. Unfortunately, TD.com makes available for
.CSV download only the last 6 statements; older ones are available only in PDF format.

Turns out there is a handy golang library called Unipdf by UniDoc that has a practical free-tier:
up to a 100 docs can be converted per month. This is way more than I will ever need.

So to use this software, you must create an account on https://cloud.unidoc.io and create an API
KEY which you then set in an environment variable called `UNIDOC_LICENSE_API_KEY`. Then simply
run `go build` and `./td-visa-pdf-xactions-to-csv path/to/statement.pdf YEAR`. The year is necessary
because the dates in the transactions do not have a year (feel free to submit a PR for getting 
the year from the statement itself). 

To do a bunch of files at once I use bash shell with

```bash
for f in path/to/statements/*.pdf; do ./td-visa-pdf-xactions-to-csv "$f"; echo; done
```

If this is useful to you, let me know!