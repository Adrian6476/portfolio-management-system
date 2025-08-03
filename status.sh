#!/bin/bash
# Portfolio Service Status Check Script

echo "📊 Portfolio Service Status Check"
echo "================================"

# Check if running from correct directory
if [ ! -f "docker-compose.yml" ]; then
    echo "🔄 Changing to portfolio project directory..."
    cd /home/adrian/repos/portfolio-management-system
fi

echo "🔍 Docker container status:"
docker compose ps 2>/dev/null || docker-compose ps 2>/dev/null || echo "Docker Compose not available"
echo ""

echo "🌐 Network connectivity test:"
echo -n "API Gateway Health Check: "
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://192.168.2.140:13001/health 2>/dev/null || echo "000")
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ OK (HTTP $HTTP_STATUS)"
elif [ "$HTTP_STATUS" = "000" ]; then
    echo "❌ Connection Failed"
else
    echo "⚠️  HTTP $HTTP_STATUS"
fi

echo -n "API Gateway WebSocket: "
# Test WebSocket endpoint using curl with upgrade headers
WS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
    http://192.168.2.140:13001/ws 2>/dev/null || echo "000")
if [ "$WS_STATUS" = "101" ]; then
    echo "✅ OK (WebSocket Upgrade $WS_STATUS)"
elif [ "$WS_STATUS" = "400" ] || [ "$WS_STATUS" = "426" ]; then
    echo "✅ Endpoint Available (HTTP $WS_STATUS - WebSocket endpoint detected)"
elif [ "$WS_STATUS" = "000" ]; then
    echo "❌ Connection Failed"
else
    echo "⚠️  HTTP $WS_STATUS (May not be WebSocket endpoint)"
fi

echo -n "Frontend Development Server: "
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://192.168.2.140:13000 2>/dev/null || echo "000")
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ OK (HTTP $HTTP_STATUS)"
elif [ "$HTTP_STATUS" = "000" ]; then
    echo "❌ Not Running"
else
    echo "⚠️  HTTP $HTTP_STATUS"
fi

echo -n "Production WebSocket (via Nginx): "
# Test production WebSocket through Nginx
WS_PROD_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
    --insecure \
    https://portfolio.adrian6476.top:8443/ws 2>/dev/null || echo "000")
if [ "$WS_PROD_STATUS" = "101" ]; then
    echo "✅ OK (WebSocket Upgrade $WS_PROD_STATUS)"
elif [ "$WS_PROD_STATUS" = "400" ] || [ "$WS_PROD_STATUS" = "426" ]; then
    echo "✅ Endpoint Available (HTTP $WS_PROD_STATUS - WebSocket endpoint detected)"
elif [ "$WS_PROD_STATUS" = "000" ]; then
    echo "❌ Connection Failed"
else
    echo "⚠️  HTTP $WS_PROD_STATUS"
fi

echo ""
echo "📈 Resource usage:"
if command -v docker &> /dev/null; then
    docker stats --no-stream portfolio_postgres portfolio_redis portfolio_nats portfolio_api_gateway 2>/dev/null || echo "No containers running"
else
    echo "Docker not available"
fi

echo ""
echo "🔗 Access URLs:"
echo "- Production Frontend: https://portfolio.adrian6476.top:8443"
echo "- Development Frontend: http://192.168.2.140:13000"
echo "- API Gateway: http://192.168.2.140:13001"
echo "- API Health: http://192.168.2.140:13001/health"
echo "- WebSocket Direct: ws://192.168.2.140:13001/ws"
echo "- WebSocket via Nginx: wss://portfolio.adrian6476.top:8443/ws"
echo "- API Documentation: http://192.168.2.140:13001/api/v1/docs (if available)"

echo ""
echo "🛠️  Quick Commands:"
echo "- Start services: ./dev.sh"
echo "- Deploy production: ./deploy.sh"
echo "- View logs: docker compose logs [service-name]"
echo "- Stop services: docker compose down"
echo "- Test WebSocket: wscat -c ws://192.168.2.140:13001/ws (if wscat installed)"
echo ""
echo "🔧 WebSocket Testing:"
echo "- Install wscat: npm install -g wscat"
echo "- Test direct: wscat -c ws://192.168.2.140:13001/ws"
echo "- Test via Nginx: wscat -c wss://portfolio.adrian6476.top:8443/ws"
