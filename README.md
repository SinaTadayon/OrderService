# Marketplace Order Service
This service manages orders based on many predefined states, starting from the creation of orders by the customers until their delivery. Of course, the delivery process is fulfilled by the sellers. It is designed for a microservice architecture.

The significant features include:
- Managing orders with MongoDB
- Managing states with composition and repository patterns
- Using the CQRS pattern with MongoDB clusters for scale-up (Designed and developed from scratch, using the FanIn/FanOut pattern)
- Using gRPC to communicate with other services
- Designing and developing a simple scheduler from scratch (using the ward/steward pattern)
- Support Docker
  
Note: Please refrain from launching the service standalone due to microservice considerations in the design.
