FROM node:22-alpine

WORKDIR /src/web

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

EXPOSE 5173

# Bind-mounted node_modules can hide the image install (or be empty on first run).
CMD ["sh", "-c", "if [ ! -x node_modules/.bin/vite ]; then npm ci; fi && exec npm run dev -- --host 0.0.0.0 --port 5173"]
