# Packages

## Purpose

host packages for distributions

## Functionality

### Implemented

* organizations
  * support for multiple orgs per server
  * organization level tokens
* distributions
  * get info
  * get versions
  * NOTE: only alpine is currently supported
* repository versions
* repositories
  * support for multiple repos per distribution/version
* architectures
* packages


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
    * 20.04 -> focal

## directory layout

### conceptual
* root
  * static
    * orgs
      * distros - public keys
        * distroversion
          * repos
            * packages (.apk/.deb/.rpm/etc)
  * config
    * orgs - pkg/repo signing private keys / tokens
      * tokens - org level tokens
      * distros - pkg/repo signing private keys / tokens
        * distroversion - pkg/repo signing private keys / tokens
          * repos - pkg/repo signing private keys / tokens

### examples
* /srv/packages
  * /static
    * /atlascloud
      * /alpine
        * /somekey.pub
        * /edge
          * /key.pub
          * /main
          * /community
        * /3.19
          * /main
      * /ubuntu
        * /repokey.pub
        * /dists
          * /24.04 -> noble
            * /main
            * /universe
            * /multiverse
            * /restricted
        * /pool
          * /main
            * /a,/b,/c... - debs
          * /multiverse
            * /a,/b,/c...
      * /fedora
        * /linux
          * /releases
            * /39
              * /Server
                * x86_64
                  * /os
                    * /Packages
                      * /a,/b,/c - rpms
  * /config
    * /atlascloud
      * /tokens
      * /alpine
        * /privkey.key
        * /tokens
          * /name
          * /2
        * /edge
          * /privkey.key
          * /main
            * /priv.key

### TODO
* server wide tokens for doing things like creating orgs, etc

