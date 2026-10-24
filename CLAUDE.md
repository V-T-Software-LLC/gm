# gm

gm is a cli tool to scaffold projects.

## Note from the Vince

I am the main author of this application. I like to keep my software simple and effective. I do not need a lot of
features or complexity in this application. Do not add complexity. Understand the problem, and apply the appropriate
fix. Do not attempt to over-engineer the problem.

I believe that data is the core driver of an application. I am taking unstructured data and presenting it to a
user for them to make decisions. Given a data type and a presentation of this data, there are only so many
ways that you can go from one to the other. I choose the simplest way.

## Glossery

How I will communicate to you

- I/We - I am referring to myself (Vince) or other potential maintainers of this software.

- You - I am referring to you, Claude Code.

- User - An individual who is using this software.

- Client - web browser or phone, wherever the user is viewing the application.

- Server - the go backend.

## Kind of projects this cli will scaffold

- web applications - the stack will usually be go backend with echo v5, postgres, redis for sessions, templ for html
views, svelte built with vite for dynamic components, htmx for communication with server from html (svelte components
will use json apis)

- cli applications - the language will will depend on the project (C, C++, Rust, or Go)

- low level projects - C or C++ projects

## Taste

- I do not believe that premature optimization is the root of all evil, but I do believe that complexity
is the root of all evil. Keep the code simple and clean.

