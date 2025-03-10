# Pulsely

Pulsely is a Go external uptime monitoring service that allows you to monitor the availability of your web applications. 
The goal of this project is to provide a way for internal services to push their status to a central monitoring service
that can be queried by external services. This is useful for monitoring the availability of internal services that may 
not be publicly accessible.

## Features
- Push status updates from internal services to a central monitoring service
- View the status of all monitored services
- Query the status of individual services
- Simple and easy to use API
- Written in Go
- Lightweight and fast
- Easy to deploy
- Supports multiple storage backends (currently only Sqlite is supported)
- Supports multiple notification channels (currently only email is supported)

