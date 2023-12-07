#!/usr/bin/env sh
set -e

cd /home/work/application

start_app() {
  gunicorn --chdir cdir app:app -w 2 --threads 2 -b 0.0.0.0:8848
}

CMD=$1

case "$CMD" in
"web")
  start_app
  ;;
*)
  exec $@
  ;;
esac
