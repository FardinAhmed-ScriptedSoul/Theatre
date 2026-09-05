<<<<<<< HEAD
# Cinema ticket booking system

Starter code for the gradual explanation on how to create a ticket booking system
to solve for the double booking system and provide a good UX for clients when booking
a ticket and avoid conflicts in a high contention enviroment.

## The Problem

Two users click "Book" on seat A1 at the same instant. Only one should win.

```bash
User A ──► read seat A1 → "free" ──► write booking ──► success
User B ──► read seat A1 → "free" ──► write booking ──► ???
```
Without any protection, both succeed. Now two people show up for the same seat.
=======
# Theatre
A simple movie booking platform .Solving Concurrency issues ....
>>>>>>> 9d4b7870f97c69d20ea15797558d9d3bafb07655
