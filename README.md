
# Form Handler

## Overview

Form Handler is a web application designed to handle form submissions securely. It includes features such as input sanitization, rate limiting, referral URL validation, CORS validation, and form field validation. This application stores form submissions in a SQLite database and allows for basic administrative actions such as viewing and deleting submissions.

## Features

- **Input Sanitization:** Ensures that form inputs are sanitized to prevent XSS and other injection attacks.
- **Rate Limiting:** Limits the number of requests a user can make to prevent abuse.
- **Referral URL Validation:** Ensures that forms are submitted from approved URLs.
- **CORS Validation:** Validates Cross-Origin Resource Sharing requests to prevent unauthorized access.
- **Form Field Validation:** Ensures that form inputs adhere to the specified rules (e.g., required fields, max length).

## Directory Structure

```
/Users/adavis/Projects/private/form-handler
├── Dockerfile
├── README.md
├── app
│   ├── backend
│   │   ├── index.html
│   │   ├── login.html
│   │   ├── rate_limits.html
│   │   └── tailwind.min.css
│   ├── config.go
│   ├── db.go
│   ├── go.mod
│   ├── go.sum
│   ├── handlers.go
│   ├── logger.go
│   ├── main.go
│   ├── middleware.go
│   ├── models.go
│   └── session.go
├── config
│   └── config.json
├── docker-compose.yml
├── examples
│   └── forms
│       ├── form_with_file.html
│       └── form_with_text.html
├── helper_scripts
│   ├── clean_and_build_container.sh
│   └── list_structure_and_content.sh
├── logs
│   └── app.log
└── tests
    ├── run_all_tests.sh
    ├── test_authentication.sh
    ├── test_cors_validation.sh
    ├── test_form_field_validation.sh
    ├── test_input_sanitization.sh
    ├── test_rate_limiting.sh
    └── test_referral_url_validation.sh
```

## Prerequisites

- Docker
- Docker Compose (optional)
- Go 1.21 or higher

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
ADMIN_USERNAME=your_admin_username
ADMIN_PASSWORD=your_admin_password
SESSION_SECRET=your_session_secret
```

These variables are used for administrative authentication and session management.

## Configuration

The application configuration is stored in `config/config.json`. Update this file with your form configurations. Example:

```json
{
    "forms": {
        "a1b2c3d4e5f6": {
            "referral_url": "http://127.0.0.1/",
            "allowed_origins": ["http://127.0.0.1"],
            "rate_limit": {
                "requests": 5,
                "duration": "1m"
            },
            "fields": [
                {"name": "name", "type": "text", "required": true, "max_length": 100},
                {"name": "email", "type": "email", "required": true, "max_length": 100},
                {"name": "message", "type": "textarea", "required": true, "max_length": 500},
                {
                    "name": "file",
                    "type": "file",
                    "required": false,
                    "max_file_size": 10485760,
                    "allowed_file_types": ["image/jpeg", "image/png", "application/pdf"]
                }
            ]
        },
        "g7h8i9j0k1l2": {
            "referral_url": "http://127.0.0.1/",
            "allowed_origins": ["http://127.0.0.1"],
            "rate_limit": {
                "requests": 5,
                "duration": "1m"
            },
            "fields": [
                {"name": "email", "type": "email", "required": true, "max_length": 100},
                {"name": "message", "type": "textarea", "required": true, "max_length": 500}
            ]
        }
    }
}
```

## Example Forms

Example HTML forms are provided in the `examples/forms` directory:

- `form_with_file.html`: A form that includes file upload.
- `form_with_text.html`: A form that includes text fields.

These examples demonstrate how to structure your forms to submit data to the Form Handler application.

## Running the Application

### Using Docker

1. **Build the Docker image:**

    ```sh
    ./helper_scripts/clean_and_build_container.sh
    ```

2. **Run the Docker container:**

    ```sh
    docker run -p 8080:8080 \
               -v "$(pwd)/app/config/config.json:/app/config/config.json" \
               -v "$(pwd)/app/logs:/app/logs" \
               -v "$(pwd)/app/uploads:/app/uploads" \
               -v "$(pwd)/app/data:/app/data" \
               --env-file "$(pwd)/.env" \
               --name form-handler \
               form-handler
    ```

### Using Docker Compose

1. **Run the application:**

    ```sh
    docker-compose up --build
    ```

### Running Locally

1. **Install dependencies:**

    ```sh
    go mod download
    ```

2. **Set environment variables:**

    ```sh
    export ADMIN_USERNAME=your_admin_username
    export ADMIN_PASSWORD=your_admin_password
    export SESSION_SECRET=your_session_secret
    ```

3. **Run the application:**

    ```sh
    go run app/main.go
    ```

## Testing

### Running Tests

Use the provided shell scripts in the `tests` directory to test various functionalities:

- **Input Sanitization:** `tests/test_input_sanitization.sh`
- **Rate Limiting:** `tests/test_rate_limiting.sh`
- **Referral URL Validation:** `tests/test_referral_url_validation.sh`
- **CORS Validation:** `tests/test_cors_validation.sh`
- **Form Field Validation:** `tests/test_form_field_validation.sh`
- **Authentication:** `tests/test_authentication.sh`

### Example

To run all tests:

```sh
tests/run_all_tests.sh
```

## License

This project is licensed under the MIT License.
