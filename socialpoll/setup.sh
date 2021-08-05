#!/usr/bin/env bash

export SP_TWITTER_KEY=5mlzNO5z7pHbjvM9tVUgK7tFk
export SP_TWITTER_SECRET=svjfHY90tQ9DC2DHTWCwDMGZAixyzVPouZ2TM8ztx6RMiY5Ake
export SP_TWITTER_ACCESSTOKEN=1411195257262198788-vfBBQmpuGtBHaipwRqcN5Toqd2qfLw
export SP_TWITTER_ACCESSSECRET=489pQJ58C71jOBBKmo1ZDRBvWMYjmzzVi7H0m2YKEWtkk

echo building twittervotes : go build ./twittervotes/ -o twittervotes
go build twittervotes/ -o twittervotes

echo start running twittervotes: ./twittervotes/twittervotes
twittervotes/twittervotes

