# Hord

Hord is a read-through/write-through caching service that sits in-front of any database.

**Why Hord?** When applications need fast access to data they keep their data in a cache. We often describe this in simple terms such as **"putting a cache in front of the database"**.  But the reality of this is not so simple. With database caches, applications have to manage writing to both a cache and a database. They have to manage what to do when data doesn't exist in the cache, and how often data should live in cache.

Hord aims to make this process much easier. It does this by giving a single service to call for reading and writing data. Hord manages what data to keep in memory, when to refresh that data and how to write data to the database.

It does this while giving the user a simple to use key/value based interface.

## Status: In Development

[![Build Status](https://travis-ci.org/madflojo/hord.svg?branch=develop)](https://travis-ci.org/madflojo/hord)

Hord is currently in development, it should not be used and expected to work... yet. This is a "for fun" project, development of this project will move at it's own pace, feel free to contribute if you want to see more.
