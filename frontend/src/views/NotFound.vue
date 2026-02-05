<template>
  <div class="not-found">
    <h1>404</h1>
    <p>Page not found</p>
    <router-link to="/" class="btn">Go Home</router-link>
  </div>
</template>

<style scoped>
.not-found {
  text-align: center;
  padding: 4rem 2rem;
}

h1 {
  font-size: 6rem;
  color: #2c3e50;
  margin-bottom: 1rem;
}

p {
  font-size: 1.5rem;
  color: #7f8c8d;
  margin-bottom: 2rem;
}

.btn {
  padding: 1rem 2rem;
  background: #3498db;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  display: inline-block;
  transition: background 0.2s;
}

.btn:hover {
  background: #2980b9;
  text-decoration: none;
}
</style>
