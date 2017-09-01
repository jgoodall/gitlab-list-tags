This is a command line tool to list tags for a gitlab repo using the [Tag API](https://docs.gitlab.com/ce/api/tags.html#list-project-repository-tags).

## Install

```sh
go get github.com/jgoodall/gitlab-list-tags
```

## Usage

To use it for any non-public repository, you must first get a `Personal access token` in your gitlab installation (save that token somewhere safe) and use the `-token` option. If your installation uses a self-signed certificate, you can use the `-insecure` option.

Your tag names must be parsable according to [semver](http://semver.org/) rules to use the `-sort-semver` option, which will print the most recent tags first. Any tag that begins with a `v` (e.g. `v1.0.0`) will have the `v` removed. When using the `-sort-semver` option, you can specify the tags to get by setting the `-since-tag` option, and all tags after the one specified will be retrieved.
