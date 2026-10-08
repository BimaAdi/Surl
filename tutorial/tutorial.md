# Surl Tutorial

Surl is a CLI HTTP client. You describe your HTTP requests once in a `surl.json`
file, then run them by key instead of re-typing long curl commands.

```sh
surl run <key>        # run one api by key
surl list             # print all available keys
surl tutorial         # print this tutorial
surl run <key> -c path/to/surl.json   # use another config file (default: surl.json)
```

---

## The `surl.json` file

A config has two top-level sections:

- `api` — a map of request keys to request specs, it can have multiple api
- `global` — values shared by every request

here a example of surl.json
```json
{
    "global": {},
    "api": {
        "ping": {
            "url": "http://localhost/ping",
        },
        "pong": {
            "url": "http://localhost/pong",
        }

    }
}
```
to run on shell just run
```sh
surl run ping
```

### api
#### Simple Get Api
```json
{
    "api": {
        "simple": {
            "url": "http://localhost",
            "method": "GET"
        }
    }
}
```

#### Query Params
```json
{
    "api": {
        "query_example": {
            "url": "http://localhost",
            "method": "GET",
            "query": {
                "page": "1",
                "page_size": "10"
            }
        }
    }
}
```
url become http://localhost/?page=1&page_size=10


#### Raw Body
```json
{
    "api": {
        "raw body": {
            "url": "http://localhost/posts",
            "method": "POST",
            "headers": {
                "content-type": "application/json"
            },
            "body": "{\"title\":\"some title\", \"body\":\"some body\"}"
        }
    }
}
```

#### Json 
```json
{
    "api": {
        "raw body": {
            "url": "http://localhost/posts",
            "method": "POST",
            "json": {
                "title": "some title",
                "body": "some body"
            }
        }
    }
}
```
automatically add header content-type: application/json

#### Form Data
```json
{
    "api": {
        "form_data_example": {
            "url": "http://localhost/posts",
            "method": "POST",
            "form": [
                { "name": "title", "value": "some title" },
                { "name": "body",  "value": "some body" }
            ]
        }
    } 
}
```
automatically set header content-type: application/x-www-form-urlencoded.
each item must contains "name" (string) and "value" (string)

#### Form Multipart
```json
{
    "api": {
        "multipart_example": {
            "url": "http://localhost/posts",
            "method": "POST",
            "form_multipart": [
                {
                  "name": "username",
                  "type": "string",
                  "value": "hello"
                },
                {
                  "name": "file",
                  "type": "file",
                  "file_name": "README.md",
                  "file_path": "./README.md"
                }
          ]
        }
    } 
}
```
automatically set header content-type: multipart/form-data .
the multi part item has two type
- string
```json
{
    "name": "<field name>",
    "type": "string",
    "value": "<field value>"
}
```
- file
```json
{
    "name": "<field name>",
    "type": "file",
    "file_name": "<file name>",
    "file_path": "<file path>"
}
```
 
### Global
Global is configuration that applied to all api.
#### variable
```json
{
  "global": {
    "variable": {
        "url": "http://127.0.0.1",
        "name": "John"
    }
  },
  "api": {
    "ping": {
        "url": "{url}/ping"
    },
    "ping_name": {
        "url": "{url}/hello?={name}"
    }
  }
}
```
variable is to make reusable string. it will replace all string with {}. 
based on the example:
- ping: http://127.0.0.1/ping
- ping_name: http://127.0.0.1/hello?=John

#### headers
```json
{
  "global": {
    "headers": {
      "Authorization": "Bearer globaltoken"
    },
  },
  "api": {
    "ping": {"url" : "http://127.0.0.1/ping"},
    "pong": {
        "url" : "http://127.0.0.1/pong",
        "headers": {
            "Authorization": "Bearer localtoken"
        }
    }
  }
}
```
it add header to all api request. the headers from local will overide global headers.
    - ping -> Authorization: Bearer globaltoken
    - pong -> Authorization: Bearer localtoken

#### Basic Auth
```json
{
  "global": {
    "basic-auth": {
      "username": "myusername",
      "password": "some-password"
    }
  },
  "api": {
    "ping": {"url" : "http://127.0.0.1/ping"},
    "pong": {
        "url" : "http://127.0.0.1/pong",
        "basic-auth": {
            "username": "myusername",
            "password": "some-password"
        }
    }
  }
}
```
it add basic-auth to all api request. Like headers local will overide global.

#### TLS Config
```json
{
  "global": {
    "tls-config": {
      "insecure-skip-verify": true
    }
  },
  "api": {}
  }
}
```
it add tls config to all api. currently only support insecure-skip-verify for other key comming soon.

