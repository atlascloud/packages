# Design Notes

Just some notes about the design of the API/etc

## code generation

Using oapi-codegen for swagger v3 (go-swagger only does v2).

## Database

This API doesn't use a standard (relational/document/graph/etc) database to store info. The info
about each repo is all stored as extended attributes of the dirs/files in the filesystem. This may
turn out to be a terrible idea.

## alpine repo layout

```txt

└── alpine
    └── edge/3.14/3.13
        ├── community
        ├── main
        │   └── x86_64
        │       ├── abuild-3.7.0_rc1-r0.apk
        │       └── APKINDEX.tar.gz
        └── testing
```

## debian repo layout

```txt
└── debian
    └── dists
        └── bullseye/bullseye-backports/testing/etc
            ├── contrib
            ├── main
            │   └── binary-amd64
            ├── non-free
```

## npm layout

