# Facebook Conversions API Token Management - Complete Guide

## 📋 What Was Changed

### New Files Created:

1. **`internal/models/facebook.go`**
   - Defines data structures for storing Facebook tokens
   - `FacebookToken` - stores the token in database
   - `FacebookTokenRequest` - for receiving token from API
   - `FacebookTokenResponse` - for sending token info (without exposing actual token)

2. **`internal/handlers/facebook.go`**
   - Handles all Facebook token operations:
     - Save/Update token
     - Get token info
     - Get token value (for API calls)
     - Delete/Deactivate token

3. **Updated `internal/server/server.go`**
   - Added Facebook handler initialization
   - Added 4 new admin routes for token management

---

## 🔧 How It Works

### Architecture Flow:

```
Facebook Dashboard → Generate Token → Your API → MongoDB Storage
                                                      ↓
                                            Use Token for Conversions API
```

### Step-by-Step Process:

1. **Token Storage:**
   - Admin saves Facebook token via API
   - Token is stored in MongoDB collection `facebook_tokens`
   - Only ONE active token at a time (old ones are deactivated)

2. **Token Security:**
   - Token is stored securely in database
   - Regular GET requests don't expose the actual token
   - Only admin can access token endpoints
   - Special endpoint `/token/value` returns actual token (for API calls)

3. **Token Management:**
   - Save new token → Creates new record, deactivates old ones
   - Update token → Updates existing active token
   - Get token → Returns metadata (not the actual token)
   - Delete token → Deactivates (soft delete)

---

## 📱 What Facebook Needs & How to Use

### Part 1: Getting Token from Facebook

**Step 1: Go to Facebook Business Manager**
1. Go to: https://business.facebook.com
2. Select your Business Account
3. Go to **Events Manager** → **Data Sources** → **Conversions API**

**Step 2: Generate Access Token**
1. Click **"Set up manually"** or **"Set up with Dataset Quality API"** (recommended)
2. Facebook will generate an access token (like the one in your screenshot)
3. Copy the token (it's long, multi-line)

**Step 3: Save Token to Your Backend**

Use your API to save it:

```bash
# 1. Login as admin first
curl -X POST http://your-server/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rifat@yopmail.com",
    "password": "rifat@123"
  }'

# Response: {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

# 2. Save Facebook token (use the token from step 1)
curl -X POST http://your-server/api/admin/facebook/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_FROM_STEP_1" \
  -H "Content-Type: application/json" \
  -d '{
    "access_token": "EAAJ6kBj5zVcBPx5N4oZBgr0tpZCdzLmJw4NnKZBt85CYi10WZASJ8E2..."
  }'
```

---

### Part 2: Using Token for Facebook Conversions API

**What is Conversions API?**
- Sends purchase/event data directly from your server to Facebook
- More reliable than browser pixel (works even with ad blockers)
- Helps Facebook track conversions better

**How to Send Events to Facebook:**

When a customer makes a purchase, send event to Facebook:

```javascript
// Example: After order is created in your backend
async function sendToFacebook(order) {
  // 1. Get Facebook token from your API
  const tokenResponse = await fetch('http://your-server/api/admin/facebook/token/value', {
    headers: {
      'Authorization': 'Bearer YOUR_ADMIN_JWT'
    }
  });
  const { access_token } = await tokenResponse.json();

  // 2. Send event to Facebook Conversions API
  const facebookResponse = await fetch(
    `https://graph.facebook.com/v18.0/YOUR_PIXEL_ID/events?access_token=${access_token}`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        data: [{
          event_name: 'Purchase',
          event_time: Math.floor(Date.now() / 1000),
          user_data: {
            em: hashEmail(order.customerEmail), // Hash email for privacy
            ph: hashPhone(order.customerPhone), // Hash phone
            fn: hashName(order.customerFirstName),
            ln: hashName(order.customerLastName),
            external_id: order.customerId
          },
          custom_data: {
            currency: 'USD',
            value: order.totalAmount,
            order_id: order.id
          }
        }]
      })
    }
  );
}
```

**Or from your Go backend:**

```go
// In your order handler, after creating order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // ... create order logic ...
    
    // Send to Facebook Conversions API
    go h.sendToFacebookConversionsAPI(order)
}

func (h *OrderHandler) sendToFacebookConversionsAPI(order *models.Order) {
    // 1. Get Facebook token
    token := h.getFacebookToken() // Call your API
    
    // 2. Prepare event data
    eventData := map[string]interface{}{
        "data": []map[string]interface{}{
            {
                "event_name": "Purchase",
                "event_time": time.Now().Unix(),
                "user_data": map[string]string{
                    "em": hashEmail(order.CustomerEmail),
                    "external_id": order.CustomerID.Hex(),
                },
                "custom_data": map[string]interface{}{
                    "currency": "USD",
                    "value": order.TotalAmount,
                    "order_id": order.ID.Hex(),
                },
            },
        },
    }
    
    // 3. Send to Facebook
    url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/events?access_token=%s", 
        "YOUR_PIXEL_ID", token)
    // Make HTTP POST request...
}
```

---

## 🔐 API Endpoints Reference

### 1. Save/Update Token
```http
POST /api/admin/facebook/token
Authorization: Bearer {admin_jwt_token}
Content-Type: application/json

{
  "access_token": "EAAJ6kBj5zVcBPx5N4oZBgr0tpZCdzLmJw4NnKZBt85CYi10WZASJ8E2...",
  "token_type": "Bearer"  // optional
}
```

**Response:**
```json
{
  "message": "Token saved successfully",
  "token": {
    "id": "6919e87154c4ce05546da897",
    "is_active": true,
    "created_by": "rifat@yopmail.com",
    "created_at": "2025-11-16T23:06:25Z",
    "updated_at": "2025-11-16T23:06:25Z"
  }
}
```

### 2. Get Token Info (Metadata Only)
```http
GET /api/admin/facebook/token
Authorization: Bearer {admin_jwt_token}
```

**Response:**
```json
{
  "token": {
    "id": "6919e87154c4ce05546da897",
    "is_active": true,
    "created_by": "rifat@yopmail.com",
    "created_at": "2025-11-16T23:06:25Z",
    "updated_at": "2025-11-16T23:06:25Z"
  }
}
```

### 3. Get Token Value (For API Calls)
```http
GET /api/admin/facebook/token/value
Authorization: Bearer {admin_jwt_token}
```

**Response:**
```json
{
  "access_token": "EAAJ6kBj5zVcBPx5N4oZBgr0tpZCdzLmJw4NnKZBt85CYi10WZASJ8E2...",
  "token_type": "Bearer",
  "is_active": true
}
```

### 4. Delete/Deactivate Token
```http
DELETE /api/admin/facebook/token
Authorization: Bearer {admin_jwt_token}
```

**Response:**
```json
{
  "message": "Token deactivated successfully"
}
```

---

## 🎯 Complete Workflow Example

### Scenario: Customer Makes a Purchase

1. **Customer completes order** → Your backend creates order
2. **Backend gets Facebook token** → Call `/api/admin/facebook/token/value`
3. **Backend sends event to Facebook** → POST to Facebook Graph API
4. **Facebook receives event** → Tracks conversion, improves ad targeting

### Benefits:
- ✅ Better ad tracking (works with ad blockers)
- ✅ More accurate conversion data
- ✅ Better ad optimization
- ✅ Improved ROI on Facebook ads

---

## 🔍 Important Notes

1. **Token Expiration:**
   - Facebook tokens can expire
   - Check token validity periodically
   - Generate new token when needed

2. **Security:**
   - Never expose token in frontend
   - Only use `/token/value` endpoint server-side
   - Store token securely in database

3. **Facebook Pixel ID:**
   - You also need your Facebook Pixel ID
   - Find it in Events Manager → Data Sources
   - Use it in Conversions API calls

4. **Event Types:**
   - Purchase
   - AddToCart
   - ViewContent
   - InitiateCheckout
   - etc.

---

## 📚 Facebook Resources

- **Conversions API Docs:** https://developers.facebook.com/docs/marketing-api/conversions-api
- **Graph API Explorer:** https://developers.facebook.com/tools/explorer/
- **Events Manager:** https://business.facebook.com/events_manager

---

## ✅ Summary

**What you have now:**
- ✅ Secure token storage in MongoDB
- ✅ Admin-only API endpoints
- ✅ Token management (save, get, delete)
- ✅ Ready to integrate with Facebook Conversions API

**Next steps:**
1. Deploy the code
2. Save your Facebook token via API
3. Integrate Conversions API calls in your order flow
4. Test with Facebook Events Manager

