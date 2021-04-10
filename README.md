# paperboy-back

[**Paperboy**](http://paperboynews.ca) is a news website that automatically aggregates and summarizes news articles from [The Guardian](https://www.theguardian.com) using [basically](https://github.com/algao1/basically). It extracts keywords, removes transition phrases and other unimportant details.

`paperboy-back` is the backend for the site, written completely in Go.

## How It Was Built

The project is deployed on GCP Kubernetes Engine.

A custom, lightweight task scheduler was written using Go's runtime-reflection to aggregate news periodically.

**MongoDB** was selected, as it allowed for easy storage, management, and querying of data with text. A custom API pagination solution was built using MongoDB. Additionally, a rudimentary search engine was also implemented using MongoDB's `searchIndex` and fuzzy matching.

**Redis** is used for server-assisted client side caching to improve performance of MongoDB queries.

## What We Learned

* Go best practices such as [application structure](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1), [error handling](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully), and [testing](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
* How to work with databases like MongoDB and Redis
* How to develop and deploy microservices using Docker and Kubernetes
* Cloud computing using GCP Kubernetes Engine and Digital Ocean

## Next Steps

* Adding unit tests and integration tests
* Migrating from GCP Kubernetes Engine to Digital Ocean
* Improving search and recommendation functionality using Lucene or ElasticSearch
* Implement a scraper to scrape other news sites