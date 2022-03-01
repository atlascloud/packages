# Packages

## Purpose

host packages for distributions

## Notes

Currently we call out to apk, so you either need to run this in an alpine container/host or have a
standalone apk binary in your path

## Functionality
### Implemented

* distributions
  * get info
  * get versions
  * NOTE: only alpine is currently supported
* organizations
  * support for multiple orgs per server
  * organization level tokens
* repos
  * support for multiple repos per org


### Planned

* distributions
  * support for more distributions
* repos
  * repo level tokens
* quotas
  * per repo/org/etc
* aliases
  * version aliases
    * 10.0 -> buster
    * 20.04 -> bionic
  *