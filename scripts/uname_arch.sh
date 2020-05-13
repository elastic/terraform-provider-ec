#!/usr/bin/env bash

arch=$(uname -m)
case $arch in
  x86_64) arch="amd64" ;;
  x86) arch="386" ;;
  i686) arch="386" ;;
  i386) arch="386" ;;
  aarch64) arch="arm64" ;;
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*) arch="armv7" ;;
esac
echo ${arch}
