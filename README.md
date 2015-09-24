# merkderwn

I write a lot of Markdown. Some of it I convert to LaTeX using MultiMarkdown, which is a little tiresome because you have to wrap all TeX commands in HTML comments, e.g.

> Quotes are great. &lt;!--\cite{Ghandi}--&gt;

"Merkdown" does little more than automatically wrap everything it identifies as LaTeX commands in HTML comments.

## Running tests

    go test *.go

Run the above in the root of the project. Requires `wdiff` to be installed on the system.
