compozis/template
=================

Compozis template provides engine for [`html/template`](https://golang.org/pkg/html/template/) construction and rendering.

## About

Features:

* support for `{{ extends "parent/template/path" }}`:
  * allows different templates to be based on different base templates
  * nesting of extended templates is supported
* decoupled from native/system filesystem:
  * allows custom templates storage (ie. built into executable)

## Decisions

- no support for automatic extension appending/postfix:
  - breaks Intellij's refactor (rename), which includes only matches containing extension,
  - better for full-text code search as it filters out matches without extension. 