# Seed DynamoDB Products and Reviews tables with sample data
# Usage: .\seed_dynamodb.ps1
# Requires: AWS CLI configured with proper credentials

$Region = "us-east-1"
$ProductsTable = "Products"
$ReviewsTable = "Reviews"

Write-Host "`nSeeding DynamoDB tables..." -ForegroundColor Green

# Helper to build DynamoDB JSON and put item
function Put-Product {
    param($id, $name, $price, $desc, $stock, $seller)
    $now = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    # Multiply price by 300 for LKR conversion
    $lkrPrice = [string]([decimal]$price * 300)
    # Use generic placeholder image
    $imageUrl = "https://via.placeholder.com/400x300/6366f1/ffffff?text=$([uri]::EscapeDataString($name))"
    
    $obj = @{
        productId = @{ S = $id }
        name = @{ S = $name }
        price = @{ N = $lkrPrice }
        description = @{ S = $desc }
        stock = @{ N = $stock }
        sellerId = @{ S = $seller }
        imageUrl = @{ S = $imageUrl }
        createdAt = @{ S = $now }
        updatedAt = @{ S = $now }
    }
    $json = ($obj | ConvertTo-Json -Compress -Depth 5)
    aws dynamodb put-item --table-name $ProductsTable --item $json --region $Region 2>$null
    Write-Host "  + $name (LKR $lkrPrice)" -ForegroundColor Green
}

function Put-Review {
    param($revId, $prodId, $text, $rating, $userId)
    $now = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $obj = @{
        reviewId = @{ S = $revId }
        productId = @{ S = $prodId }
        text = @{ S = $text }
        rating = @{ N = $rating }
        userId = @{ S = $userId }
        createdAt = @{ S = $now }
    }
    $json = ($obj | ConvertTo-Json -Compress -Depth 5)
    aws dynamodb put-item --table-name $ReviewsTable --item $json --region $Region 2>$null
    Write-Host "  + Review for $prodId" -ForegroundColor Green
}

Write-Host "`nAdding products..." -ForegroundColor Cyan

Put-Product "prod-001" "Wireless Bluetooth Headphones" "59.99" "Premium wireless headphones with noise cancellation, 30-hour battery life, and crystal-clear audio." "150" "seller-demo-001"
Put-Product "prod-002" "USB-C Fast Charger 65W" "29.99" "Compact GaN charger supporting PD 3.0 and QC 4.0. Compatible with laptops, tablets and phones." "300" "seller-demo-001"
Put-Product "prod-003" "Mechanical Gaming Keyboard" "89.99" "RGB mechanical keyboard with Cherry MX Blue switches, aluminium frame and customisable macros." "75" "seller-demo-001"
Put-Product "prod-004" "4K Ultra HD Webcam" "119.99" "Professional 4K webcam with auto-focus, low-light correction and built-in privacy shutter." "50" "seller-demo-002"
Put-Product "prod-005" "Portable SSD 1TB" "79.99" "Pocket-sized solid state drive with USB 3.2 Gen 2 speeds up to 1050 MB/s and shock resistance." "200" "seller-demo-002"
Put-Product "prod-006" "Smart Home Security Camera" "49.99" "1080p indoor security camera with night vision, two-way audio and motion detection alerts." "120" "seller-demo-002"
Put-Product "prod-007" "Ergonomic Office Chair" "249.99" "Adjustable lumbar support, breathable mesh back, 4D armrests and smooth-rolling casters." "30" "seller-demo-001"
Put-Product "prod-008" "Stainless Steel Water Bottle 750ml" "24.99" "Double-wall vacuum insulated bottle keeps drinks cold for 24h or hot for 12h. BPA free." "500" "seller-demo-003"
Put-Product "prod-009" "Wireless Charging Pad" "19.99" "Qi-certified 15W fast wireless charger with LED indicator and non-slip surface." "250" "seller-demo-003"
Put-Product "prod-010" "Noise-Cancelling Earbuds" "129.99" "True wireless earbuds with active noise cancellation, transparency mode, and IPX5 water resistance." "100" "seller-demo-001"
Put-Product "prod-011" "LED Desk Lamp" "34.99" "Dimmable LED desk lamp with 5 colour temperatures, USB charging port and flexible gooseneck." "180" "seller-demo-003"
Put-Product "prod-012" "Laptop Stand Adjustable" "39.99" "Aluminium laptop riser with 6 height levels, foldable design. Fits 10-17 inch laptops." "90" "seller-demo-002"

Write-Host "`nAdding reviews..." -ForegroundColor Cyan

Put-Review "rev-001" "prod-001" "Amazing sound quality and great battery life!" "5" "user-demo-001"
Put-Review "rev-002" "prod-001" "Good headphones but noise cancellation could be better." "4" "user-demo-002"
Put-Review "rev-003" "prod-002" "Charges my laptop super fast. Very compact too!" "5" "user-demo-001"
Put-Review "rev-004" "prod-003" "Love the clicky keys! Perfect for gaming." "5" "user-demo-003"
Put-Review "rev-005" "prod-004" "Crystal clear video in meetings. Auto-focus works great." "4" "user-demo-002"
Put-Review "rev-006" "prod-005" "Fast transfer speeds and very portable. Worth every penny." "5" "user-demo-001"
Put-Review "rev-007" "prod-006" "Easy setup and clear night vision. App is decent." "4" "user-demo-003"
Put-Review "rev-008" "prod-007" "Best chair I have ever owned. Back pain is gone!" "5" "user-demo-002"
Put-Review "rev-009" "prod-008" "Keeps water cold all day. Great build quality." "4" "user-demo-001"
Put-Review "rev-010" "prod-010" "Incredible noise cancellation for the price!" "5" "user-demo-003"

Write-Host "`nSeeding complete! 12 products, 10 reviews." -ForegroundColor Green
