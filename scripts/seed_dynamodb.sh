#!/bin/bash
# Seed DynamoDB Products and Reviews tables with sample data
# Usage: ./seed_dynamodb.sh
# Requires: AWS CLI configured with proper credentials

REGION="us-east-1"
PRODUCTS_TABLE="Products"
REVIEWS_TABLE="Reviews"

echo "🌱 Seeding DynamoDB tables..."

# ─────────────────────────────────────────────────────────────────
# Products
# ─────────────────────────────────────────────────────────────────

echo "📦 Adding products..."

declare -a PRODUCTS=(
  '{"productId":{"S":"prod-001"},"name":{"S":"Wireless Bluetooth Headphones"},"price":{"N":"59.99"},"description":{"S":"Premium wireless headphones with noise cancellation, 30-hour battery life, and crystal-clear audio."},"stock":{"N":"150"},"sellerId":{"S":"seller-demo-001"},"createdAt":{"S":"2024-01-15T10:00:00Z"},"updatedAt":{"S":"2024-01-15T10:00:00Z"}}'
  '{"productId":{"S":"prod-002"},"name":{"S":"USB-C Fast Charger 65W"},"price":{"N":"29.99"},"description":{"S":"Compact GaN charger supporting PD 3.0 and QC 4.0. Compatible with laptops, tablets and phones."},"stock":{"N":"300"},"sellerId":{"S":"seller-demo-001"},"createdAt":{"S":"2024-01-16T10:00:00Z"},"updatedAt":{"S":"2024-01-16T10:00:00Z"}}'
  '{"productId":{"S":"prod-003"},"name":{"S":"Mechanical Gaming Keyboard"},"price":{"N":"89.99"},"description":{"S":"RGB mechanical keyboard with Cherry MX Blue switches, aluminium frame and customisable macros."},"stock":{"N":"75"},"sellerId":{"S":"seller-demo-001"},"createdAt":{"S":"2024-01-17T10:00:00Z"},"updatedAt":{"S":"2024-01-17T10:00:00Z"}}'
  '{"productId":{"S":"prod-004"},"name":{"S":"4K Ultra HD Webcam"},"price":{"N":"119.99"},"description":{"S":"Professional 4K webcam with auto-focus, low-light correction and built-in privacy shutter."},"stock":{"N":"50"},"sellerId":{"S":"seller-demo-002"},"createdAt":{"S":"2024-02-01T10:00:00Z"},"updatedAt":{"S":"2024-02-01T10:00:00Z"}}'
  '{"productId":{"S":"prod-005"},"name":{"S":"Portable SSD 1TB"},"price":{"N":"79.99"},"description":{"S":"Pocket-sized solid state drive with USB 3.2 Gen 2 speeds up to 1050 MB/s and shock resistance."},"stock":{"N":"200"},"sellerId":{"S":"seller-demo-002"},"createdAt":{"S":"2024-02-05T10:00:00Z"},"updatedAt":{"S":"2024-02-05T10:00:00Z"}}'
  '{"productId":{"S":"prod-006"},"name":{"S":"Smart Home Security Camera"},"price":{"N":"49.99"},"description":{"S":"1080p indoor security camera with night vision, two-way audio and motion detection alerts."},"stock":{"N":"120"},"sellerId":{"S":"seller-demo-002"},"createdAt":{"S":"2024-02-10T10:00:00Z"},"updatedAt":{"S":"2024-02-10T10:00:00Z"}}'
  '{"productId":{"S":"prod-007"},"name":{"S":"Ergonomic Office Chair"},"price":{"N":"249.99"},"description":{"S":"Adjustable lumbar support, breathable mesh back, 4D armrests and smooth-rolling casters."},"stock":{"N":"30"},"sellerId":{"S":"seller-demo-001"},"createdAt":{"S":"2024-03-01T10:00:00Z"},"updatedAt":{"S":"2024-03-01T10:00:00Z"}}'
  '{"productId":{"S":"prod-008"},"name":{"S":"Stainless Steel Water Bottle 750ml"},"price":{"N":"24.99"},"description":{"S":"Double-wall vacuum insulated bottle keeps drinks cold for 24h or hot for 12h. BPA free."},"stock":{"N":"500"},"sellerId":{"S":"seller-demo-003"},"createdAt":{"S":"2024-03-10T10:00:00Z"},"updatedAt":{"S":"2024-03-10T10:00:00Z"}}'
  '{"productId":{"S":"prod-009"},"name":{"S":"Wireless Charging Pad"},"price":{"N":"19.99"},"description":{"S":"Qi-certified 15W fast wireless charger with LED indicator and non-slip surface."},"stock":{"N":"250"},"sellerId":{"S":"seller-demo-003"},"createdAt":{"S":"2024-03-15T10:00:00Z"},"updatedAt":{"S":"2024-03-15T10:00:00Z"}}'
  '{"productId":{"S":"prod-010"},"name":{"S":"Noise-Cancelling Earbuds"},"price":{"N":"129.99"},"description":{"S":"True wireless earbuds with active noise cancellation, transparency mode, and IPX5 water resistance."},"stock":{"N":"100"},"sellerId":{"S":"seller-demo-001"},"createdAt":{"S":"2024-04-01T10:00:00Z"},"updatedAt":{"S":"2024-04-01T10:00:00Z"}}'
  '{"productId":{"S":"prod-011"},"name":{"S":"LED Desk Lamp"},"price":{"N":"34.99"},"description":{"S":"Dimmable LED desk lamp with 5 colour temperatures, USB charging port and flexible gooseneck."},"stock":{"N":"180"},"sellerId":{"S":"seller-demo-003"},"createdAt":{"S":"2024-04-05T10:00:00Z"},"updatedAt":{"S":"2024-04-05T10:00:00Z"}}'
  '{"productId":{"S":"prod-012"},"name":{"S":"Laptop Stand Adjustable"},"price":{"N":"39.99"},"description":{"S":"Aluminium laptop riser with 6 height levels, foldable design. Fits 10-17 inch laptops."},"stock":{"N":"90"},"sellerId":{"S":"seller-demo-002"},"createdAt":{"S":"2024-04-10T10:00:00Z"},"updatedAt":{"S":"2024-04-10T10:00:00Z"}}'
)

for PRODUCT in "${PRODUCTS[@]}"; do
  aws dynamodb put-item \
    --table-name "$PRODUCTS_TABLE" \
    --item "$PRODUCT" \
    --region "$REGION" 2>/dev/null && echo "  ✅ Added product" || echo "  ❌ Failed to add product"
done

# ─────────────────────────────────────────────────────────────────
# Reviews
# ─────────────────────────────────────────────────────────────────

echo "⭐ Adding reviews..."

declare -a REVIEWS=(
  '{"reviewId":{"S":"rev-001"},"productId":{"S":"prod-001"},"text":{"S":"Amazing sound quality and great battery life!"},"rating":{"N":"5"},"userId":{"S":"user-demo-001"},"createdAt":{"S":"2024-02-01T09:00:00Z"}}'
  '{"reviewId":{"S":"rev-002"},"productId":{"S":"prod-001"},"text":{"S":"Good headphones but noise cancellation could be better."},"rating":{"N":"4"},"userId":{"S":"user-demo-002"},"createdAt":{"S":"2024-02-15T14:00:00Z"}}'
  '{"reviewId":{"S":"rev-003"},"productId":{"S":"prod-002"},"text":{"S":"Charges my laptop super fast. Very compact too!"},"rating":{"N":"5"},"userId":{"S":"user-demo-001"},"createdAt":{"S":"2024-02-20T11:00:00Z"}}'
  '{"reviewId":{"S":"rev-004"},"productId":{"S":"prod-003"},"text":{"S":"Love the clicky keys! Perfect for gaming."},"rating":{"N":"5"},"userId":{"S":"user-demo-003"},"createdAt":{"S":"2024-03-05T16:00:00Z"}}'
  '{"reviewId":{"S":"rev-005"},"productId":{"S":"prod-004"},"text":{"S":"Crystal clear video in meetings. Auto-focus works great."},"rating":{"N":"4"},"userId":{"S":"user-demo-002"},"createdAt":{"S":"2024-03-10T10:00:00Z"}}'
  '{"reviewId":{"S":"rev-006"},"productId":{"S":"prod-005"},"text":{"S":"Fast transfer speeds and very portable. Worth every penny."},"rating":{"N":"5"},"userId":{"S":"user-demo-001"},"createdAt":{"S":"2024-03-15T09:00:00Z"}}'
  '{"reviewId":{"S":"rev-007"},"productId":{"S":"prod-006"},"text":{"S":"Easy setup and clear night vision. App is decent."},"rating":{"N":"4"},"userId":{"S":"user-demo-003"},"createdAt":{"S":"2024-03-20T13:00:00Z"}}'
  '{"reviewId":{"S":"rev-008"},"productId":{"S":"prod-007"},"text":{"S":"Best chair I have ever owned. Back pain is gone!"},"rating":{"N":"5"},"userId":{"S":"user-demo-002"},"createdAt":{"S":"2024-04-01T15:00:00Z"}}'
  '{"reviewId":{"S":"rev-009"},"productId":{"S":"prod-008"},"text":{"S":"Keeps water cold all day. Great build quality."},"rating":{"N":"4"},"userId":{"S":"user-demo-001"},"createdAt":{"S":"2024-04-05T08:00:00Z"}}'
  '{"reviewId":{"S":"rev-010"},"productId":{"S":"prod-010"},"text":{"S":"Incredible noise cancellation for the price!"},"rating":{"N":"5"},"userId":{"S":"user-demo-003"},"createdAt":{"S":"2024-04-15T11:00:00Z"}}'
)

for REVIEW in "${REVIEWS[@]}"; do
  aws dynamodb put-item \
    --table-name "$REVIEWS_TABLE" \
    --item "$REVIEW" \
    --region "$REGION" 2>/dev/null && echo "  ✅ Added review" || echo "  ❌ Failed to add review"
done

echo ""
echo "✅ Seeding complete!"
echo "   Products: ${#PRODUCTS[@]}"
echo "   Reviews:  ${#REVIEWS[@]}"
