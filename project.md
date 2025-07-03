# Help Desk Portal Backend - Project Specification

## Project Overview

Create a HIPAA-compliant help desk portal backend API using Go and MSSQL, designed for software/technology and healthcare industries. The system will support end users, support agents, administrators, and managers with comprehensive ticket management and real-time communication features.

## Technology Stack

- **Framework**: Go Echo web framework
- **Database**: SQLite
- **Authentication**: JWT tokens with secure token management
- **ORM**: GORM for database operations
- **Validation**: Go validator or custom validation
- **Testing**: Go testing package with testify
- **Documentation**: Swagger/OpenAPI with go-swagger
- **File Upload**: Multipart form handling with security validation
- **Email**: SMTP integration for notifications
- **Real-time**: WebSocket implementation with Gorilla WebSocket
- **Deployment**: Azure Cloud Services
- **Security**: HIPAA-compliant encryption and audit logging

## Core Features

### 1. User Management (HIPAA Compliant)

- User registration and authentication with JWT
- Role-based access control (End User, Support Agent, Administrator, Manager)
- User profiles with contact information
- Password reset functionality with secure token generation
- Email verification for account activation
- HIPAA-compliant user data handling and encryption
- Audit logging for all user actions

### 2. Ticket Management

- Create, read, update, delete tickets with proper authorization
- Ticket status tracking (Open, In Progress, Resolved, Closed)
- Priority levels (Low, Medium, High, Critical)
- Categories and tags for organization
- File attachments support with virus scanning
- Ticket assignment to agents with workload balancing
- Ticket history and audit trail for compliance
- Escalation rules based on priority and response time
- HIPAA-compliant data handling for sensitive information

### 3. Communication System

- Internal notes between agents (not visible to end users)
- Public comments visible to users
- Email notifications for status changes and assignments
- Real-time updates via WebSocket for live ticket updates
- Ticket escalation system with automated notifications
- Secure messaging with encryption for sensitive data

### 4. Reporting & Analytics

- Ticket statistics and metrics
- Agent performance tracking
- Response time analytics
- Customer satisfaction metrics
- HIPAA-compliant reporting with data anonymization
- Export functionality with secure data handling

## Database Schema (MSSQL)

### Users Table

```sql
CREATE TABLE Users (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Email NVARCHAR(255) UNIQUE NOT NULL,
    PasswordHash NVARCHAR(255) NOT NULL,
    FirstName NVARCHAR(100) NOT NULL,
    LastName NVARCHAR(100) NOT NULL,
    Role NVARCHAR(20) NOT NULL CHECK (Role IN ('END_USER', 'SUPPORT_AGENT', 'ADMINISTRATOR', 'MANAGER')),
    IsVerified BIT DEFAULT 0,
    IsActive BIT DEFAULT 1,
    LastLoginAt DATETIME2,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2 DEFAULT GETDATE(),
    CreatedBy UNIQUEIDENTIFIER,
    UpdatedBy UNIQUEIDENTIFIER
);
```

### Tickets Table

```sql
CREATE TABLE Tickets (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Title NVARCHAR(255) NOT NULL,
    Description NVARCHAR(MAX) NOT NULL,
    Status NVARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (Status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED')),
    Priority NVARCHAR(20) NOT NULL DEFAULT 'MEDIUM' CHECK (Priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    CategoryID UNIQUEIDENTIFIER,
    AssignedAgentID UNIQUEIDENTIFIER,
    CreatedByID UNIQUEIDENTIFIER NOT NULL,
    EscalatedAt DATETIME2,
    EscalatedTo UNIQUEIDENTIFIER,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2 DEFAULT GETDATE(),
    ResolvedAt DATETIME2,
    DueDate DATETIME2,
    FOREIGN KEY (CategoryID) REFERENCES Categories(ID),
    FOREIGN KEY (AssignedAgentID) REFERENCES Users(ID),
    FOREIGN KEY (CreatedByID) REFERENCES Users(ID),
    FOREIGN KEY (EscalatedTo) REFERENCES Users(ID)
);
```

### Categories Table

```sql
CREATE TABLE Categories (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    Name NVARCHAR(100) NOT NULL,
    Description NVARCHAR(500),
    ParentID UNIQUEIDENTIFIER,
    IsActive BIT DEFAULT 1,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (ParentID) REFERENCES Categories(ID)
);
```

### Comments Table

```sql
CREATE TABLE Comments (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    TicketID UNIQUEIDENTIFIER NOT NULL,
    UserID UNIQUEIDENTIFIER NOT NULL,
    Content NVARCHAR(MAX) NOT NULL,
    IsInternal BIT DEFAULT 0,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    UpdatedAt DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (TicketID) REFERENCES Tickets(ID),
    FOREIGN KEY (UserID) REFERENCES Users(ID)
);
```

### Attachments Table

```sql
CREATE TABLE Attachments (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    TicketID UNIQUEIDENTIFIER NOT NULL,
    Filename NVARCHAR(255) NOT NULL,
    FilePath NVARCHAR(500) NOT NULL,
    FileSize BIGINT NOT NULL,
    MimeType NVARCHAR(100) NOT NULL,
    UploadedByID UNIQUEIDENTIFIER NOT NULL,
    IsVirusScanned BIT DEFAULT 0,
    IsSafe BIT DEFAULT 0,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (TicketID) REFERENCES Tickets(ID),
    FOREIGN KEY (UploadedByID) REFERENCES Users(ID)
);
```

### AuditLog Table (HIPAA Compliance)

```sql
CREATE TABLE AuditLog (
    ID UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    UserID UNIQUEIDENTIFIER,
    Action NVARCHAR(100) NOT NULL,
    TableName NVARCHAR(100) NOT NULL,
    RecordID UNIQUEIDENTIFIER,
    OldValues NVARCHAR(MAX),
    NewValues NVARCHAR(MAX),
    IPAddress NVARCHAR(45),
    UserAgent NVARCHAR(500),
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (UserID) REFERENCES Users(ID)
);
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password
- `POST /api/v1/auth/verify-email` - Email verification

### Users

- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `GET /api/v1/users/agents` - Get all agents (Admin/Manager only)
- `PUT /api/v1/users/:id/role` - Update user role (Admin only)
- `GET /api/v1/users/:id/audit-log` - Get user audit log (Admin only)

### Tickets

- `GET /api/v1/tickets` - Get tickets (with filtering, pagination, sorting)
- `POST /api/v1/tickets` - Create new ticket
- `GET /api/v1/tickets/:id` - Get ticket details
- `PUT /api/v1/tickets/:id` - Update ticket
- `DELETE /api/v1/tickets/:id` - Delete ticket (Admin only)
- `POST /api/v1/tickets/:id/assign` - Assign ticket to agent
- `POST /api/v1/tickets/:id/status` - Update ticket status
- `GET /api/v1/tickets/:id/history` - Get ticket history
- `POST /api/v1/tickets/:id/escalate` - Escalate ticket

### Comments

- `GET /api/v1/tickets/:id/comments` - Get ticket comments
- `POST /api/v1/tickets/:id/comments` - Add comment
- `PUT /api/v1/comments/:id` - Update comment
- `DELETE /api/v1/comments/:id` - Delete comment

### Attachments

- `POST /api/v1/tickets/:id/attachments` - Upload attachment
- `GET /api/v1/attachments/:id` - Download attachment
- `DELETE /api/v1/attachments/:id` - Delete attachment

### Categories

- `GET /api/v1/categories` - Get all categories
- `POST /api/v1/categories` - Create category (Admin only)
- `PUT /api/v1/categories/:id` - Update category (Admin only)
- `DELETE /api/v1/categories/:id` - Delete category (Admin only)

### Analytics

- `GET /api/v1/analytics/overview` - Get overview statistics
- `GET /api/v1/analytics/tickets` - Get ticket analytics
- `GET /api/v1/analytics/agents` - Get agent performance
- `GET /api/v1/analytics/export` - Export data (HIPAA compliant)

## Implementation Steps

### Phase 1: Project Setup & HIPAA Foundation

- [x] Initialize Go project with proper module structure
- [x] Set up database with HIPAA-compliant configurations (SQLite implemented, MSSQL planned)
- [x] Configure GORM for database operations
- [x] Set up environment variables and configuration management
- [x] Create basic folder structure following Go best practices
- [x] Configure middleware (CORS, logging, security headers)
- [x] Implement HIPAA-compliant audit logging system

### Phase 2: Authentication & User Management

- [x] Implement JWT authentication with secure token management
- [x] Create user registration and login endpoints
- [x] Set up password hashing with bcrypt
- [x] Implement role-based middleware for authorization
- [x] Add email verification system structure
- [x] Create password reset functionality
- [x] Implement HIPAA-compliant user data handling

### Phase 3: Core Ticket System

- [x] Set up database models and migrations for tickets, categories, comments, attachments
- [x] Implement CRUD operations for tickets
- [x] Add ticket status management with workflow validation
- [x] Create ticket assignment system with workload balancing
- [x] Implement priority and category management
- [ ] Add file upload functionality with virus scanning
- [x] Implement escalation rules and automation

### Phase 4: Communication Features

- [ ] Implement comment system with internal/external visibility
- [ ] Add internal notes functionality for agents
- [ ] Set up email notifications with templates
- [ ] Configure WebSocket for real-time updates
- [ ] Implement ticket escalation with notifications
- [ ] Add secure messaging for sensitive data

### Phase 5: Advanced Features & Analytics

- [ ] Create reporting and analytics endpoints
- [ ] Implement search and filtering with full-text search
- [ ] Add pagination and sorting for large datasets
- [ ] Create data export functionality with HIPAA compliance
- [ ] Set up comprehensive audit logging
- [ ] Implement performance monitoring and metrics

### Phase 6: Testing & Documentation

- [ ] Write unit tests for all business logic
- [ ] Create integration tests for API endpoints
- [ ] Set up API documentation with Swagger
- [ ] Perform security testing and HIPAA compliance validation
- [ ] Create deployment configuration for Azure
- [ ] Set up monitoring and alerting

## HIPAA Compliance Requirements

### Data Protection

- All PHI (Protected Health Information) must be encrypted at rest and in transit
- Implement AES-256 encryption for sensitive data
- Use TLS 1.3 for all API communications
- Secure file storage with encrypted attachments

### Access Control

- Role-based access control with least privilege principle
- Multi-factor authentication for administrative access
- Session management with automatic timeout
- IP whitelisting for administrative access

### Audit Logging

- Comprehensive audit trail for all data access and modifications
- Log user authentication attempts and failures
- Track all PHI access with user identification
- Maintain audit logs for minimum 6 years

### Data Handling

- Data anonymization for reporting and analytics
- Secure data disposal procedures
- Backup encryption and secure storage
- Incident response procedures for data breaches

## Azure Deployment Configuration

### Infrastructure

- Azure App Service for API hosting
- Azure SQL Database for MSSQL
- Azure Blob Storage for file attachments
- Azure Key Vault for secrets management
- Azure Application Insights for monitoring

### Security

- Azure Active Directory integration
- Network Security Groups for access control
- Azure Firewall for additional protection
- Managed Identity for secure service communication

### Scaling

- Auto-scaling based on CPU and memory usage
- Database scaling with Azure SQL elastic pools
- CDN for static content delivery
- Load balancing for high availability

## Success Criteria

- All endpoints return proper HTTP status codes
- Authentication and authorization work correctly
- File uploads are secure and virus-scanned
- Real-time updates work via WebSocket
- Email notifications are sent appropriately
- Analytics provide meaningful insights
- API documentation is complete and accurate
- Test coverage exceeds 80%
- Performance meets requirements under load
- HIPAA compliance validation passes
- Azure deployment is successful and secure

## Risk Mitigation

- Regular security audits and penetration testing
- Automated vulnerability scanning
- HIPAA compliance monitoring and reporting
- Disaster recovery procedures
- Regular backup testing and validation
- Incident response plan and procedures
