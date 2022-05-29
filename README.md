# configurefs

confugurefs is a read-only fuse filesystem which uses [Golang template](https://pkg.go.dev/text/template) and a [YAML](https://yaml.org/) variables file to generate files on the fly.

You can use it for example to template your configuration files and automatically generate them during `stat` or `read`.
So instead of the need of something like [ansible](https://www.ansible.com/), you can simply change the variables on your system and re-read the configuration file to get the new values.

It was implemented to give a base for simple configurations tools, which now only have to change variables and restart services.

Run `make` to build and take a look at `make setup_test` and `./configurefs -help` to learn how to use it.
