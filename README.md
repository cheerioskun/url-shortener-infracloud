# A Simple URL Shortener

## Disclaimer
This is 100% human-made. With <3, Hemant.

## Usage
The server can be run from source using go > 1.23
```bash
> go run ./... & # Or run it in a different terminal
>
> # An example full flow from shorten to redirect
> curl -X POST http://localhost:3000/short -d '{"URL":"https://everything.curl.dev/http/redirects.html"}' | jq .URL | xargs curl -vv
*   Trying 127.0.0.1:3000...
* Connected to localhost (127.0.0.1) port 3000 (#0)
> GET /long/tibW5bAiIudRembjhyjPMd7AMPKuYzRGDiUogVIEyIk= HTTP/1.1
> Host: localhost:3000
> User-Agent: curl/7.81.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 307 Temporary Redirect
< Location: https://everything.curl.dev/http/redirects.html
< Date: Mon, 21 Jul 2025 19:43:28 GMT
< Content-Length: 0
<
* Connection #0 to host localhost left intact
```
You can also run it as a container
```bash
docker run -p 3000:3000 cheerioskun/shortener:stripped
```


## Task and Requirements
The task at hand is to create an API with the functionality of shortening long URLs into small shareable links which redirect to the orginal.
More formally:
1. It can take in a long valid URL and return a deterministic short valid URL.
2. It can take a short valid URL and return the original long URL if present or redirect to a static 404 page if not present.
3. It maintains some metrics like frequency of domains for the URLs shortened.

## Architecture
We will keep it super simple. 
1. The state/db will be in memory: a simple map. We will make this a sync.Map to make the APIs possibly concurrent.
2. We will have deterministic hash function that reduces the long string (the URL) to a fixed size. We can select something like BLAKE2 for hashing and base64 for encoding.
3. We will have some parsing logic for the URLs coming in. This will help us 
    3.1 Validate that the input is a valid URL
    3.2 Aggregate metrics on domain name. We will keep this simple, no need for openmetrics format right now.

The API can look something like: 
- POST /short - Takes in a body with URL as parameter. Returns a single string back: a valid shortened URL.
- GET /long/{blurb} - Takes in the blurb/encodedString in the query path. Returns a redirect code and the correctly set Location header.
- GET /metrics - Returns the top 3 shortened domain names.

### Framework
We'll be using Gin for the REST API.

### Side Note on Collision Risk
Since we are mapping from an infinite domain of strings to only a 5 character blurb, it is important to analyse the collision risk between two shortened URLs.
Considering our hash function a perfect hash function, i.e. it uniformly distributes from our biased domain to the range of 256 bits (64 bytes), our problem reduces to considering
the set of all possible 5 character combinations in base64. This is simply 64^5, about 10^9. 
Referring to the [birthday problem](https://betterexplained.com/articles/understanding-the-birthday-paradox/), this means that we can add about sqrt(10^9) => 30,000 keys before
the probability of a collision is 50%. This seems like a good enough number, which we can easily change later, or implement a TTL to flush out our existing keys.