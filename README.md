# qtemp

qtemp is a caching layer for the Golang HTML Template engine.

It heavily relies on the idea that you have one master-file, and one or more
template file which provide content for that master file.

Nested templates are not really supported (yet).

Clearing cache on runtime is not supported (yet) either.
