gobranchdocs is a tool meant to automate the process of getting pkg.go.dev module 
documentation for whatever the head of the current checked out branch of a given 
project is.  This is meant to automate this process:

https://github.com/golang/go/issues/36811#issuecomment-579404726

Quick and dirty hack, but figured it might be useful for others.  This works
for whatever the current checked out branch of the package is, provided it's pushed
to the remote and pkg.go.dev and proxy.golang.org can reach the remote.  Upon 
success, it will attempt to open the appropriate pkg.go.dev package documentation
page in the default browser.
