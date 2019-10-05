# Komment

Minimalist comment solution for static webpages.

## Design goals
- Easy to install and setup
- Easy to integrate into own page-layout
- Maximize compliance with GDPR by only storing what is really necessary
- Can be used as server or via CGI
- **No dependencies** client-side while using Komment (e.g. jQuery)
- **No dependencies** during run-time (No specific versions of a script environment or DB servers ...)
- **No dependencies** during build-time (Go standard library only)

## Installation

1)	Build (or use a pre-built binary for) Komment for your OS/Architecture
	and copy it - together with the `templates` and `config` directories -
	to your server.
	(**CGI**: make sure it is world-accessible and has
	its execution permissions set correctly.)
	
2)	[Configure](#configuration) Komment via `komment.json` in the `config`-
	subdirectory.
	(**CGI**: make sure the config file is not accessible
	from the webserver. The `config`-directory in this repo contains an
	appropriate `.htaccess` file for Apache.)
	
3)	Modify the HTML [templates](#templates) in the `template`-directory to match your
	sites layout.
	
4)	Make the [necessary modifications](#usage) to your HTML pages / theme.

5)	(**CGI**: Done.)
	Configure your webserver to make Komment accessible to the outside world.
	
## Configuration

`$(EXECUTABLE)/config/komment.json`

```jsonc
{
	// BASIC SETTINGS
	// ==============
	// World-accessible URI of Komment.
	// Despite the name, this is also the URI of Komment when
	// running as server. Either directly or any proxy-mapping
	// happening through nginx, Apache &c.
	"CgiPath": "/komment/komment.cgi",
	// Instead of Komment running as a one-off via CGI it can
	// also run as a server. Set ListenOn to the host:port it
	// should ListenOn. e.g. "0.0.0.0:1234".
	"ListenOn": "",
  
  	// MESSAGE STORAGE & RESTRICTIONS
	// ==============================
	// either an absolute path or relative path to the executable
	// where messages will be stored
	"MessagesPath": "messages",
	// number of seconds after posting during which a message can be edited
	"EditWindow": 60,
	// maximum length of a message.
	// any longer and the message will be truncated and "..." will be added.
	// if MaxLength is not set or 0 no limit will be imposed
	"MaxLength": 100000,
	// maximum length of the poster's name
	// if MaxNameLength is not set or 0 no limit will be imposed
	"MaxNameLength": 40,
	
	// LOOK & FEEL
	// ===========
	// absolute path or relative to executable where templates are located
	"TemplatePath": "templates",
	// format used during output.
	// reference date: Mon Jan 2 15:04:05 -0700 MST 2006
	"DateFormat": "2006-Jan-2, Mon 15:04 MST",
	
	// ID VALIDATION
	// =============
	// absolute or relative (to executable) path to a text file
	// with each line being a regular expression to match an id against.
	// if any match is found the valid is interpreted to be valid.
	// set Whitelist to "", to disable it.
	"Whitelist": "config/id-whitelist.txt",
	// if the id is not whitelisted (or Whitelist being disabled)
	// Komment uses IdValidator to verify the validity of an ID.
	// IdValidator is the path to an executable which is called
	// with 2 arguments:
	// - raw-komment-id
	// - komment-id
	// if IdValidator runs successfully and returns 0 the id is
	// interpreted as valid.
	"IdValidator": "/home/johndoe/bin/id_validator",
	
 	// RATE LIMITING
	// =============
	// Does not work when running via CGI.
	// Limiting is implemented as token bucket, with each message
	// consuming one token.
	//
	// Maximum number of tokens in bucket.
	"BucketSize": 5,
	// Number of seconds until a new token is generated.
	"TokenRate": 60,
	
	// MAIL NOTIFICATION
	// =================
	// Get notification of new comments sent to SmtpTo.
	// If SmtpTo is empty, no notifications will be sent.
	"SmtpTo": "admin@example.com"
	"SmtpFrom": "comment-notification@example.com",
	"SmtpHostname": "mail.example.com",
	"SmtpPort": 587
	"SmtpUser": "example",
	"SmtpPassword": "Staple1HorseToTheHay!",
}
```

## Integration

### Templates

Komment uses 3 templates to tailor the HTML output to your needs:
- count.html.tmpl
- form.html.tmpl
- message.html.tmpl

### Usage

Add the following snippets to your webpage.

Count of messages for a certain thread (**data-komment-id**):
```html
<div class="komment_count" data-komment-id="example-2"></div>
````

Insert a form to add a new comment:
```html
<div class="komment_form" data-komment-id="example-2"></div>
```

Render all messages:
```html
<div class="komment_messages" data-komment-id="example-2">Enable Javascript to see comments</div>
```

At the end of your website add:
```html
<script src="/komment/komment.cgi?r=script"></script>
<script>Komment.init()</script>
```
