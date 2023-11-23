All contributions are welcome! Please read the following before you start.

Keep in mind that these are not set in stone and can be changed or modified anytime.

## Goals

These are a list of my goals for developing the Grape library. Each and every contribution must abide by them.

1. Having zero third-party dependencies.
2. Grape is a wrapper around Go standard library. Hence, all APIs must be compatible with it.
3. Simplicity is key. Grape isn't trying to do more than it was intended to.
4. Breaking changes may occur prior to reaching v1.
5. No breaking changes in each major release. I don't see releasing v2 or beyond, but if the need for breaking changes
   arises, they'll be included in the next major version.
6. If a feature that already exists in Grape gets added to the standard library, it'll be removed or repurposed to keep
   compatibility.

## Conventions

Your contributions must follow these conventions:

1. No breaking changes, unless you have a *really* good reason for it.
2. Code must be clean and clear with minimal indent level, nesting and complications.
3. Put comments as much as it's necessary.
4. Functions and types must be kept private in most cases.
5. Export only if you have to, otherwise avoid it.
6. If something is exported, it needs documentation as well.
7. Tests and examples must be updated as well, if applicable.
8. Before adding a feature, ask yourself: "Is this something I expect to see in the standard library?" If you believe it
   is, then proceed.
9. Keep git commits small and isolated; with a clear message stating what happened and starting with a capital verb.
10. Each commit must be compilable on its own.

## Notice

I may edit your contributions, modify your code and make changes to it before merging.

While I appreciate all the feedbacks, pull requests and inputs, there's no guaranty for them to be implemented. Please
don't take it the wrong way, as I may be happy with the project as it stands.
