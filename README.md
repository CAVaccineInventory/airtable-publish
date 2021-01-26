# airtable-export

## How This Works

`airtable-export` is a worker that periodically fetches data from
Airtable, runs it through a sanitization pass (including but not
limited to removing superfluous or sensitive keys), then uploads the
results.

`monitoring` is a black-box monitoring of the output of that, which is
used to page on failures.  One is deployed for each of the four output
products (prod/staging times counties/locations).
