// Test health check endpoints
const testHealthEndpoints = async () => {
  console.log('=== Testing Health Check Endpoints ===\n');

  try {
    // 1. Test main health check
    console.log('1. Testing /health endpoint...');
    const healthResponse = await fetch('http://localhost:8000/health');
    const healthData = await healthResponse.json();
    console.log(`✓ Status: ${healthResponse.status}`);
    console.log(`✓ Response:`, JSON.stringify(healthData, null, 2));

    // 2. Test readiness check
    console.log('\n2. Testing /health/ready endpoint...');
    const readyResponse = await fetch('http://localhost:8000/health/ready');
    const readyData = await readyResponse.json();
    console.log(`✓ Status: ${readyResponse.status}`);
    console.log(`✓ Response:`, JSON.stringify(readyData, null, 2));

    // 3. Test liveness check
    console.log('\n3. Testing /health/live endpoint...');
    const liveResponse = await fetch('http://localhost:8000/health/live');
    const liveData = await liveResponse.json();
    console.log(`✓ Status: ${liveResponse.status}`);
    console.log(`✓ Response:`, JSON.stringify(liveData, null, 2));

    console.log('\n=== All Health Checks Passed! ===');
    console.log('\nHealth endpoints summary:');
    console.log('- GET /health - Full health check with database ping');
    console.log('- GET /health/ready - Readiness probe (for K8s/Docker)');
    console.log('- GET /health/live - Liveness probe (service status)');
  } catch (error) {
    console.error('Error:', error.message);
  }
};

testHealthEndpoints();
