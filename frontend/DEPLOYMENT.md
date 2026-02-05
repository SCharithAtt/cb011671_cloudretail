# CloudRetail Frontend - Deployment Guide

## AWS S3 + CloudFront Deployment

### Prerequisites

- AWS CLI configured
- S3 bucket created
- CloudFront distribution created

### Build and Deploy

```bash
# 1. Update environment variables for production
# Edit .env.production with your API Gateway URL

# 2. Build for production
npm run build

# 3. Upload to S3
aws s3 sync dist/ s3://your-cloudretail-bucket --delete

# 4. Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id YOUR_DISTRIBUTION_ID --paths "/*"
```

### S3 Bucket Configuration

Enable static website hosting:
- Index document: `index.html`
- Error document: `index.html` (for SPA routing)

Bucket policy (replace with your bucket name):
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-cloudretail-bucket/*"
    }
  ]
}
```

### CloudFront Configuration

1. **Origin Settings**:
   - Origin Domain: `your-cloudretail-bucket.s3.amazonaws.com`
   - Origin Access: Public or OAI

2. **Default Cache Behavior**:
   - Viewer Protocol Policy: Redirect HTTP to HTTPS
   - Allowed HTTP Methods: GET, HEAD, OPTIONS
   - Cache Policy: CachingOptimized

3. **Custom Error Responses**:
   - HTTP Error Code: 404
   - Customize Error Response: Yes
   - Response Page Path: `/index.html`
   - HTTP Response Code: 200

4. **SSL/TLS Certificate**:
   - Use AWS Certificate Manager (ACM) for custom domain

---

## Docker Deployment

### Build Docker Image

```bash
# Build image
docker build -t cloudretail-frontend:latest .

# Tag for registry
docker tag cloudretail-frontend:latest your-registry/cloudretail-frontend:latest

# Push to registry
docker push your-registry/cloudretail-frontend:latest
```

### Run with Docker

```bash
# Run locally
docker run -p 80:80 cloudretail-frontend:latest

# Run with environment file
docker run -p 80:80 --env-file .env.production cloudretail-frontend:latest
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  frontend:
    image: cloudretail-frontend:latest
    ports:
      - "80:80"
    environment:
      - VITE_API_GATEWAY_URL=https://api.cloudretail.example.com
      - VITE_GRAPHQL_URL=https://api.cloudretail.example.com/product/graphql
    restart: unless-stopped
```

---

## Kubernetes Deployment

### Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudretail-frontend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cloudretail-frontend
  template:
    metadata:
      labels:
        app: cloudretail-frontend
    spec:
      containers:
      - name: frontend
        image: your-registry/cloudretail-frontend:latest
        ports:
        - containerPort: 80
        env:
        - name: VITE_API_GATEWAY_URL
          value: "https://api.cloudretail.example.com"
        - name: VITE_GRAPHQL_URL
          value: "https://api.cloudretail.example.com/product/graphql"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: cloudretail-frontend
  namespace: default
spec:
  type: LoadBalancer
  selector:
    app: cloudretail-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
```

### Apply to Kubernetes

```bash
kubectl apply -f deployment.yaml

# Check deployment
kubectl get deployments
kubectl get pods
kubectl get services

# Get external IP
kubectl get service cloudretail-frontend
```

---

## Environment Configuration

### Production Environment Variables

Create `.env.production` with production values:

```bash
# API Gateway endpoint
VITE_API_GATEWAY_URL=https://api.cloudretail.example.com

# GraphQL endpoint
VITE_GRAPHQL_URL=https://api.cloudretail.example.com/product/graphql
```

### Build-time vs Runtime Variables

**Important**: Vite embeds environment variables at build time.

For runtime configuration:
1. Use a configuration file loaded at runtime
2. Or rebuild for each environment
3. Or use server-side injection (e.g., Nginx sub_filter)

---

## CI/CD Pipeline Example

### GitHub Actions

`.github/workflows/deploy.yml`:

```yaml
name: Deploy Frontend

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Install dependencies
      run: npm ci
      working-directory: ./frontend
    
    - name: Build
      run: npm run build
      working-directory: ./frontend
      env:
        VITE_API_GATEWAY_URL: ${{ secrets.API_GATEWAY_URL }}
        VITE_GRAPHQL_URL: ${{ secrets.GRAPHQL_URL }}
    
    - name: Deploy to S3
      uses: jakejarvis/s3-sync-action@master
      with:
        args: --delete
      env:
        AWS_S3_BUCKET: ${{ secrets.AWS_S3_BUCKET }}
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        AWS_REGION: 'us-east-1'
        SOURCE_DIR: 'frontend/dist'
    
    - name: Invalidate CloudFront
      uses: chetan/invalidate-cloudfront-action@v2
      env:
        DISTRIBUTION: ${{ secrets.CLOUDFRONT_DISTRIBUTION_ID }}
        PATHS: '/*'
        AWS_REGION: 'us-east-1'
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

---

## Performance Optimization

### Build Optimization

Already configured in `vite.config.ts`:
- Code splitting
- Tree shaking
- Minification
- Asset optimization

### Additional Optimizations

1. **Enable Gzip/Brotli**:
   - CloudFront: Enable compression
   - Nginx: Already configured in nginx.conf

2. **Caching Strategy**:
   - Static assets: Long cache (1 year)
   - index.html: No cache or short cache

3. **CDN Configuration**:
   - CloudFront or comparable CDN
   - Edge locations closer to users

---

## Monitoring

### CloudWatch (AWS)

Monitor CloudFront metrics:
- Requests
- Bytes transferred
- Error rate (4xx, 5xx)

### Application Monitoring

Add application monitoring (e.g., Sentry):

```typescript
// main.ts
import * as Sentry from "@sentry/vue";

Sentry.init({
  app,
  dsn: "YOUR_SENTRY_DSN",
  integrations: [
    new Sentry.BrowserTracing({
      routingInstrumentation: Sentry.vueRouterInstrumentation(router),
    }),
  ],
  tracesSampleRate: 1.0,
});
```

---

## Security Considerations

1. **HTTPS Only**: Always use HTTPS in production
2. **Environment Variables**: Never commit secrets to version control
3. **CORS**: Backend must allow frontend domain
4. **CSP Headers**: Consider adding Content Security Policy
5. **API Keys**: Never embed API keys in frontend code

---

## Rollback Strategy

### S3 + CloudFront

```bash
# Keep previous versions in S3
# Enable versioning on S3 bucket

# Rollback: restore previous version
aws s3 sync s3://your-bucket-backup/ s3://your-bucket/ --delete

# Invalidate CloudFront
aws cloudfront create-invalidation --distribution-id YOUR_ID --paths "/*"
```

### Kubernetes

```bash
# Rollback to previous deployment
kubectl rollout undo deployment/cloudretail-frontend

# Check rollout status
kubectl rollout status deployment/cloudretail-frontend
```

---

## Troubleshooting

### Build Errors

- Check Node.js version (18+)
- Clear `node_modules` and reinstall
- Check for TypeScript errors: `npm run typecheck`

### Deployment Issues

- Verify environment variables are set
- Check CloudFront distribution settings
- Verify S3 bucket permissions
- Check CloudFront invalidation status

### Runtime Errors

- Check browser console for errors
- Verify API endpoints are reachable
- Check CORS configuration on backend
- Verify JWT tokens are valid
