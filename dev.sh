#!/bin/bash
# Portfolio Development Environment Startup Script

echo "🔧 Starting Portfolio development environment"

# Change to project directory
cd /home/adrian/repos/portfolio-management-system

# 1. Start backend services with development overrides
echo "🚀 Starting backend services (development mode)..."
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 2. Wait for services to be ready
echo "⏳ Waiting for backend services..."
sleep 10

# 3. Check backend status
echo "📊 Backend service status:"
docker compose ps

echo ""
echo "💻 To start frontend development server:"
echo "cd frontend && pnpm dev"
echo ""
echo "🌐 Access URLs:"
echo "- Frontend Dev: http://192.168.2.140:13000"
echo "- API Gateway: http://192.168.2.140:13001"
echo "- PostgreSQL: localhost:15432 (dev access)"
echo "- Redis: localhost:16379 (dev access)"
echo "- NATS: localhost:14222 (dev access)"
echo "- NATS Monitor: http://localhost:18222 (dev access)"
echo "- Production: https://portfolio.adrian6476.top:8443"
echo ""
echo "🔧 Environment Configuration:"
echo "- Development uses: http://192.168.2.140:13001/api/v1"
echo "- Production uses: https://portfolio.adrian6476.top:8443/api/v1"
