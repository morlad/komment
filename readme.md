# Komment

Minimalist comment script for static webpages.

## Design goals
- Easy to install and setup
- No dependencies client-side while using Komment (i.e. no jQuery)
- No dependencies during run-time (No specific versions of a script environment or DB servers ...)
- No dependencies during build-time (Go standard library only)
- Easy to integrate into own page-layout
- Compliant with GDPR

## Installation

## Configuration

`$(EXECUTABLE)/config/komment.json`

```json
{
  "CgiPath": "/komment/komment.cgi",
	"ListenOn": "",
  
  /* message storage & restrictions */
	"MessagesPath": "messages",
	"EditWindow": 60,
	"MaxLength": 100000,
	"MaxNameLength": 30,
	
	/* look & feel */
	"TemplatePath": "templates",
	"DateFormat": "2006-Jan-2, Mon 15:04 MST",
	
	/* id validation */
	"Whitelist": "whitelist.txt",
	"IdValidator": "/home/johndoe/bin/id_validator",
	
  /* rate limiting */
	"BucketSize": 5,
	"TokenRate": 60,
	
	/* mail on new comment */
	"SmtpTo": "admin@example.com"
	"SmtpFrom": "comment-notification@example.com",
	"SmtpHostname": "mail.example.com",
	"SmtpPort": 587
	"SmtpUser": "example",
	"SmtpPassword": "Staple1HorseToTheHay!",
}
```

### Basic Settings

### Look & Feel

### Message Storage & Restrictions

### ID Validation

### Rate Limiting
Rate Limiting does not work when running via CGI.

Limiting is implemented as [token bucket](https://en.wikipedia.org/wiki/Token_bucket)s per IP address.

- BucketSize
  Maximum number of tokens in a bucket.
- TokenRate
  Number of seconds until a new token is generated.

### Mail

## Integration
