# router-config-client

Client application to configure customer's router of an ISP provider. 
This client open a localhost web-socket channel to talk to a web interface where the feedback is displayed to the end-user. 
Once the configuration is done by customer, the client try to detected the router on local network and do the upload of the 
configuration file retrieved from a configuration API.
