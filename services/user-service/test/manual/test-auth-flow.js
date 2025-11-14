// Test Authentication Flow
// Run this after starting the server with npm run start:dev

const API_BASE = 'http://localhost:8081/api/v1';

// Test Data
const testUser = {
  name: 'Auth Test User',
  email: `authtest${Date.now()}@example.com`,
  password: 'SecurePass123!',
  push_token: 'test-push-token',
  preferences: {
    email: true,
    push: true,
  },
};

async function testAuthFlow() {
  console.log('Testing Authentication Flow\n');

  try {
    // Step 1: Register a new user (public endpoint)
    console.log('Creating new user (public endpoint)...');
    const registerResponse = await fetch(`${API_BASE}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(testUser),
    });

    const registerData = await registerResponse.json();
    console.log('User created:', registerData);
    console.log();

    if (!registerData.success) {
      console.error('Failed to create user');
      return;
    }

    const userId = registerData.data.user_id;

    // Step 2: Login with credentials
    console.log('Logging in...');
    const loginResponse = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: testUser.email,
        password: testUser.password,
      }),
    });

    const loginData = await loginResponse.json();
    console.log('Login successful:', loginData);
    console.log();

    if (!loginData.success) {
      console.error('Failed to login');
      return;
    }

    const token = loginData.data.access_token;

    // Step 3: Try accessing protected endpoint WITHOUT token (should fail)
    console.log(
      '3️Accessing protected endpoint WITHOUT token (should fail)...',
    );
    const unauthorizedResponse = await fetch(`${API_BASE}/users`, {
      method: 'GET',
    });

    const unauthorizedData = await unauthorizedResponse.json();
    console.log(
      `Status: ${unauthorizedResponse.status}`,
      unauthorizedData.message || unauthorizedData,
    );
    console.log();

    // Step 4: Access protected endpoint WITH valid token (should succeed)
    console.log('Accessing protected endpoint WITH token (should succeed)...');
    const authorizedResponse = await fetch(`${API_BASE}/users`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const authorizedData = await authorizedResponse.json();
    console.log('Protected endpoint accessed:', authorizedData);
    console.log();

    // Step 5: Get specific user preferences with token
    console.log('Getting user preferences with token...');
    const preferencesResponse = await fetch(
      `${API_BASE}/users/${userId}/preferences`,
      {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      },
    );

    const preferencesData = await preferencesResponse.json();
    console.log('User preferences:', preferencesData);
    console.log();

    // Step 6: Try with invalid token (should fail)
    console.log(' Trying with INVALID token (should fail)...');
    const invalidTokenResponse = await fetch(`${API_BASE}/users`, {
      method: 'GET',
      headers: {
        Authorization: 'Bearer invalid-token-12345',
      },
    });

    const invalidTokenData = await invalidTokenResponse.json();
    console.log(
      `Status: ${invalidTokenResponse.status}`,
      invalidTokenData.message || invalidTokenData,
    );
    console.log();

    console.log('All authentication tests completed!');
  } catch (error) {
    console.error('Test failed:', error.message);
  }
}

// Run the test
testAuthFlow();
