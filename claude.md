
This is my saga pattern for the order service of my ticket selling system.

Saga Pattern Implementation:

The Order Service implements the Saga Orchestration Pattern to manage distributed transactions across Event, Inventory, and Payment services. Each order creation is a saga with multiple steps:

Saga Steps (Forward Flow):

Validate Event - Verify event exists and is published (Event Service gRPC)
Validate Event Config - Check event configuration (Event Service gRPC)
Validate Checkout Token - If waitroom enabled, verify JWT token
Get Ticket Classes - Fetch ticket information (Inventory Service gRPC)
Check Availability - Verify sufficient ticket inventory (Inventory Service gRPC)
Reserve Tickets - Atomically reserve tickets with 15-min hold (Inventory Service gRPC) ✓ Saga tracking begins
Create Order - Persist order record (MongoDB) ✓ Saga tracked
Create Order Items - Persist order line items (MongoDB) ✓ Saga tracked
Create Payment Intent - Generate payment URL (Payment Service gRPC)
Return Payment URL - Complete order creation
Compensation/Rollback (Reverse Flow): If any step fails after saga tracking begins, the service automatically compensates:

Release Tickets - Cancel inventory reservation (Inventory Service gRPC)
Delete Order Items - Remove order line items (MongoDB)
Delete Order - Remove order record (MongoDB)

