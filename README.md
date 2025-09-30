# Pizza shop

To run this thing first create a `.env` file:

```
DB_USER=root
DB_PASS=mypassword
DB_PEPPER=some_random_string
# Optional: set to 1/true to wipe and reseed dev DB once on startup
# Leave unset (recommended) so data persists across restarts
# DB_RESET=1
```

Run the backend from the repo root:

```
go run ./backend
```

Notes on database initialization:
- By default, the server will only create tables if missing and seed data if empty.
- Your changes (admin-created pizzas/ingredients, user registrations) will now persist across restarts.
- If you need a clean slate for development, temporarily set `DB_RESET=1` when starting the server:

```
DB_RESET=1 go run ./backend
```

Then run normally again (without DB_RESET) so changes persist.



## TODO:

- [x] my sql setup
- [x] Login/register
- [ ] Accounts
    - [x] Add a view account button (go to /account)
    - [x] Add a log out button
    - [ ] Admin login screen
- [x] Menu
    - [x] Menu HTML prototype
    - [x] Menu CRUD (pizza, ingredient, etc)
    - [x] Display menu on client
    - [x] Display vegan/vegetarian
    - [x] Calculate price
- [ ] Shopping cart
    - [ ] Cart HTML
    - [ ] Functionality to add pizza to cart
- [ ] Ordering
    - [ ] Order CRUD
    - [ ] Ability to confirm order
    - [ ] Order confirmation screen
    - [ ] Transactions, rollback
- [ ] Delivery person
    - [ ] Delivery person CRUD
    - [ ] Delivery person login
    - [ ] Delivery person view assigned order
    - [ ] Delivery person mark order as delivered/failed/etc
- [ ] Desserts & drinks
    - [ ] Extra item CRUD
    - [ ] Add to menu
    - [ ] Ordering desserts and drinks
- [ ] Discount codes
    - [ ] Discount code CRUD
    - [ ] Admin create discount codes
    - [ ] Birthday discount code
    - [ ] Applying discount code in order
- [ ] Testing
    - [ ] Code for generating sample orders and sample accounts
- [ ] Reports
    - [ ] Undelivered orders
    - [ ] Top 3 pizzas sold in the past month
    - [ ] Earning reports filtered by X



## Optional TODO
- [ ] Add a delete account button
- [ ] Add updating account details