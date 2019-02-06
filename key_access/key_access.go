package key_access

func GetKey() []byte {
	// for the record, I hate this
	return []byte(`{
		"type": "service_account",
		"project_id": "reborne",
		"private_key_id": "dd707e5a0edef8fb87fcd57d17e281eb9ae2e849",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC3upRZt6+WjW8l\nRPvFrGCT5nFUb3dMSTTEPvUNeIx/jppKHfnBxo9zdv/5J2Z09wZSCiqQrv/q5MmU\n4WYKvzu5a8lLj5yzZ28PT1FqO1+zlu2rMVs8Mz+P1Zkb+k2FF3ZwvXEnNTgcAJEK\nyDXZkhL3NFokpe6F3EcmF9e8og0Fh55kMa/HEMa6l8oS9kX4fZJkU2hV0uUcNyRD\n81oVFS6dQuKndHAbXv4C2ul+cDg1oXyQ3s/RKlT+LVBZHlHTaqDc7csLfahGCHha\nXNm/YjgbneRsRldj1pqCm1zmy6bmGYrMmBX2/W+CxuQMOYQUXCyMDQiyWVj7tfBh\nP+7IeV9hAgMBAAECggEAFVg8AWNNZyZ1xeTqghfJXY1dV7OebrKninAM/JxnaC9L\nPfaXDDvS8RVfXhUFemuzZIcRVF88VI/xWlZISlHWsK80ws5DpHUNCkCLAxY0Q1My\nt9uDT65dgUqNn9o3tNFZvXXIxkcQxnILZ68EUIs/oFAI8+6CFDOX4XVOJzhFXoGk\nn1NeYbPCr5lMdUS3SbNI5Vyh171CfpiDRu0jYjSvaPDDj2hdluAO32OAsWpjCHG5\nPW1C/hMCiD/wClEeGLZo3XkZTPnHl1XM/Xj8DFcEE2NdVu22NZWtrcmkXFkhLXNM\nc3eGiyPSHtjy2XUj2sxtfzav1BoAoSv21sjNA1v8FQKBgQD//u8tQRgN2FA2Bj/b\nN3Yxx/RLJDJoBeK8qV1QdYMmrhKBWlAHJe8Izi1tOdPq2mD6xrRcdnDDlF0asPcB\nDbJ02n57OkR3K84EWZ3Txa31EIAwwVWtW+YA/IiwKsPt909kRd/Ubqd5RPKTfvBj\nS7dbZYEWvYRu2zqddIAVkOFDFQKBgQC3u1goBgvxFOq73aroUsZ7JKVp5KJjjhzc\n9+KpyyTeaeRD4AdIqe5bScODr0lUbnNhod4+9Tc/pjxMi7plPwyW1b39XKGzDY5S\n+b62LRy1FJDyW2kh0jalKJbTkeALwuMAYUfpWYURYNqr16+m7SHrbk9A5xxSpc/l\nD3pHtwsuHQKBgGgzKL/O4y+fgOa8nHlqld2lejarwSi+XJBWj/kUHBI+gKHOVQzT\nz/xRkAQJqczKnvb0sq2AOF4jodIffisbnCwcU6dtDDlFx1HV+Hwe1rQNx2AREgLC\niViVcj3i6mWOaO5z0qvxbpHaErMe1FJWm4fERUswURueeLlmlkww8MARAoGBAJaZ\nqrYL1sgCxDG/jfKmvuh7blbxQKZn+4KocZOJ3yusEp4MSQwntif/u5H6IRpi+pKh\nksF6UJIMmcqIkf2hg5kzlGrT/fr9dpbO/aLoMWrAc9skUHWXkJEqRw8euE4LrfRG\nySId7bQD9tn6jpE+OJp5Ld9eUNnx7gms+SdFg5WFAoGASIjqCGXDwIJwhkBC2d+M\n3O2tPTHLjrg3xhJ2zbloWKRSjxOxZiFgUgkzodOwclVR6TWXQ7CXKxhW3J3Ups+R\n7R7a58ZVjkY7bXAmKJBvXpPEDBeTeLu9nxnH4KQjZjB7y3mM8bRuHKv0eVlp6rRG\nSKXjP4dbrLQsiZW6kb5cinY=\n-----END PRIVATE KEY-----\n",
		"client_email": "speechliason@reborne.iam.gserviceaccount.com",
		"client_id": "114240450579356253479",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/speechliason%40reborne.iam.gserviceaccount.com"
	}`)
}
