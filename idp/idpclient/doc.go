/*
Package idpclient provides a client for the IdentityProvider-App

An idpclient with sensible defaults can be created as follows:

	c,_ := idpclient.New()


If you don't want to use the defaults provide one or more options:

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	c, _ := idpclient.New(idpclient.HttpClient(httpClient))
*/
package idpclient
