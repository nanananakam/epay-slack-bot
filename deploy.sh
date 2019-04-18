#!/usr/bin/env bash

heroku container:push web && heroku container:release web
# heroku open
# heroku config:set ENV_VAR="環境変数を指定" --app "アプリ名を指定"
