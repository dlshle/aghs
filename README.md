# Another Go HTTP Server
This is another Go HTTP Server with support for service management and a built-in simple trie-tree based parametrized path handler.

## How to use?
Please check the `example` directory under the root dir. It contains a simple HTTP Server with a simple Student service for CRUD student records. It also has a simple authentication middleware for update/delete requests.

## What's so special about this HTTP server?
Another Go HTTP Server itself does not use any third-party library at all(contrib package contains adapters for some 3rd party libraries though). It's compact and fast with support for 2 types of path parameters(paramed path parameter `:xxx` and wildcard parameter `*yyy`).
