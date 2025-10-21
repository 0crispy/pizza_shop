# Test Data Generator

This directory contains utilities for generating test data.

## Generate Test Data

To populate your database with sample customers, delivery people, and orders:

```bash
cd tools
go run generate_test_data.go
```

This will create:
- **20 customers** with random names, addresses, and details
- **5 delivery people** 
- **30 orders** with random pizzas and extras

All test accounts use the password: `password123`

### Sample Usernames
- Customers: `AlexSmith123`, `JamieJones456`, etc.
- Delivery: `delivery_AlexSmith789`, etc.

### Prerequisites
Make sure you have:
1. Database initialized with pizzas and ingredients
2. Server running (or at least database accessible)
3. Extra items (desserts/drinks) in the database (optional)

The generator will automatically create realistic orders with:
- 1-4 pizzas per order
- Random quantities (1-3 of each pizza)
- 60% chance of including desserts/drinks
- Random order statuses (IN_PROGRESS, DELIVERED, FAILED)
- Order dates spread across the last 90 days
