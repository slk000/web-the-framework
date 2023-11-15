package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	Param      *param
	urlQueries url.Values // Cache the url queries
}

// BindJSON fills val with JSON data
func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("body is nil")
	}
	if val == nil {
		return errors.New("val is nil")
	}
	decoder := json.NewDecoder(c.Req.Body)
	//decoder.DisallowUnknownFields() // do not allow unknown fields in JSON
	//decoder.UseNumber() // Use `Number`(string) as the type of numbers
	return decoder.Decode(val)
}

// FormValue gets value of `key` in form data
func (c *Context) FormValue(key string) (string, error) {
	// parsing multiple times is ok
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}
	// c.Req.Form = params in POST, PUT, PATCH and URL
	// c.Req.PostForm = params in POST, PUT, PATCH body
	val := c.Req.FormValue(key)
	return val, nil
}

/*
func (c *Context) BindForm(val any) error {

}
*/

// QueryValue gets a query param `key`
func (c *Context) QueryValue(key string) string {
	if c.urlQueries == nil {
		c.urlQueries = c.Req.URL.Query()
	}
	// if not present, return empty string is fine
	return c.urlQueries.Get(key)
}

// PathValue gets the value of path param `key`
func (c *Context) PathValue(key string) string {
	return (*c.Param)[key]
}
