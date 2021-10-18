#!/usr/bin/env bash

arch=$(uname -m)
case $arch in
  x86_64) arch="x86_64" ;;
  x86) arch="i386" ;;
  i686) arch="i386" ;;
  i386) arch="i386" ;;
  aarch64) arch="arm64" ;;
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*) arch="armv7" ;;
esac
echo ${arch}
