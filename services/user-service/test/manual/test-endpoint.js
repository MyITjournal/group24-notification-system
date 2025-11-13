// Simple test script to debug the endpoint
const testEndpoint = async () => {
  try {
    const response = await fetch('http://localhost:8000/api/v1/users', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: 'Second User',
        email: 'test' + Date.now() + '@example.com',
        password: 'password123',
        preferences: {
          email: true,
          push: false,
        },
      }),
    });

    console.log('Status:', response.status);
    console.log('Headers:', Object.fromEntries(response.headers.entries()));

    const text = await response.text();
    console.log('Raw Response:', text);

    try {
      const json = JSON.parse(text);
      console.log('JSON Response:', json);
    } catch (e) {
      console.log('Response is not JSON');
    }
  } catch (error) {
    console.error('Error:', error.message);
    console.error('Full error:', error);
  }
};

testEndpoint();
