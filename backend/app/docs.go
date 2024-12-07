package main

// @title CashOne API
// @version 1.0
// @description API Server for CashOne expense tracking application

// @contact.name API Support
// @contact.url https://github.com/Lotarcc/cashone
// @contact.email semyon.kolesnikov@outlook.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @x-extension-openapi {"example": "value on a json format"}

// @tag.name health
// @tag.description Health check endpoints for monitoring service status

// @tag.name auth
// @tag.description Authentication and authorization endpoints

// @tag.name cards
// @tag.description Card management endpoints for both manual and Monobank cards

// @tag.name transactions
// @tag.description Transaction management endpoints for tracking expenses and income

// @tag.name categories
// @tag.description Category management endpoints for organizing transactions

// @tag.name monobank
// @tag.description Monobank integration endpoints for syncing cards and transactions
