#!/bin/bash
# Portfolio Project Deployment Script

echo "🚀 Portfolio Project Deployment Script"

# Change to project directory
cd /home/adrian/repos/portfolio-management-system

# 1. Validate environment
echo "🔍 Validating environment..."
if [ ! -f "frontend/.env.production" ]; then
    echo "❌ Missing frontend/.env.production file"
    exit 1
fi

# 2. Build frontend for production
echo "📦 Building frontend application..."
cd frontend
NODE_ENV=production pnpm build
if [ $? -ne 0 ]; then
    echo "❌ Frontend build failed"
    exit 1
fi
cd ..

# 3. Start backend services
echo "🔧 Starting backend services..."
docker compose up -d

# 4. Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 15

# 5. Check service status
echo "✅ Checking service status..."
docker compose ps

# 6. Test deployment
echo "🧪 Testing deployment..."
echo -n "Backend Health: "
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:13001/health 2>/dev/null || echo "000")
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ OK"
else
    echo "❌ Failed (HTTP $HTTP_STATUS)"
fi

echo "🎉 Deployment complete!"
echo "🌐 Frontend access: https://portfolio.adrian6476.top:8443"
echo "🔗 API access: https://portfolio.adrian6476.top:8443/api/v1"
echo ""
echo "📋 Service Status:"
echo "- API Gateway: http://192.168.2.140:13001"
echo "- PostgreSQL: Internal network only"
echo "- Redis: Internal network only"
echo "- NATS: Internal network only"
echo ""
echo "To check logs: docker-compose logs [service-name]"
echo "To stop services: docker-compose down"
