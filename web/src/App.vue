<script setup>
import { computed, onMounted, ref } from 'vue'

const connections = ref([])
const selectedID = ref('')
const catalog = ref(null)
const query = ref('')
const databasesLoading = ref(true)
const loading = ref(false)
const error = ref('')
let catalogRequest = 0

const databases = computed(() => connections.value.flatMap((connection) => connection.databases))
const selectedDatabase = computed(() => databases.value.find((database) => database.id === selectedID.value))
const tables = computed(() => {
  if (!catalog.value) return []
  const normalized = query.value.trim().toLowerCase()
  if (!normalized) return catalog.value.tables
  return catalog.value.tables.filter((table) => JSON.stringify(table).toLowerCase().includes(normalized))
})

async function selectDatabase() {
  catalogRequest += 1
  const requestID = catalogRequest
  if (!selectedID.value) {
    catalog.value = null
    loading.value = false
    error.value = ''
    return
  }
  const databaseID = selectedID.value
  loading.value = true
  error.value = ''
  catalog.value = null
  try {
    const response = await fetch(`/api/databases/${encodeURIComponent(databaseID)}/catalog`)
    const payload = await response.json()
    if (requestID !== catalogRequest) return
    if (!response.ok) throw new Error(payload.error || `catalog request failed: ${response.status}`)
    catalog.value = payload
  } catch (err) {
    if (requestID !== catalogRequest) return
    error.value = err.message
  } finally {
    if (requestID === catalogRequest) loading.value = false
  }
}

onMounted(async () => {
  databasesLoading.value = true
  try {
    const response = await fetch('/api/databases')
    if (!response.ok) throw new Error(`database list request failed: ${response.status}`)
    connections.value = await response.json()
    databasesLoading.value = false
    if (databases.value.length === 1) {
      selectedID.value = databases.value[0].id
      await selectDatabase()
    }
  } catch (err) {
    error.value = err.message
  } finally {
    databasesLoading.value = false
  }
})
</script>

<template>
  <main class="shell">
    <header class="hero">
      <div>
        <p class="eyebrow">Data Dictionary</p>
        <h1>{{ catalog?.name || 'DB Library' }}</h1>
        <p class="subtitle">Choose a database to inspect its live schema metadata.</p>
      </div>
      <div class="controls">
        <label class="database-select">
          <span>Database</span>
          <select v-model="selectedID" @change="selectDatabase">
            <option value="">Select a database…</option>
            <optgroup v-for="connection in connections" :key="connection.id" :label="`${connection.name} · ${connection.engine}`">
              <option v-for="database in connection.databases" :key="database.id" :value="database.id">
                {{ database.name }} · {{ database.environment || 'unclassified' }}
              </option>
            </optgroup>
          </select>
        </label>
        <label v-if="catalog" class="search">
          <span>Search catalog</span>
          <input v-model="query" placeholder="tables, columns, descriptions…" />
        </label>
      </div>
    </header>

    <p v-if="error" class="error">{{ error }}</p>
    <div v-else-if="databasesLoading || loading" class="loading-panel" role="status" aria-live="polite">
      <p class="loading-label">{{ loading ? 'Loading catalog…' : 'Loading databases…' }}</p>
      <p v-if="loading && selectedDatabase" class="loading-detail">{{ selectedDatabase.name }}</p>
      <div class="progress" aria-hidden="true">
        <div class="progress-bar"></div>
      </div>
    </div>
    <p v-else-if="!databases.length" class="empty">No databases configured. Copy config/databases.example.json to config/databases.json to get started.</p>
    <p v-else-if="!catalog" class="empty">Select a database to browse its schema.</p>
    <template v-else>
      <div class="summary">
        <span>{{ tables.length }} tables</span>
        <span>{{ tables.reduce((count, table) => count + table.columns.length, 0) }} columns</span>
        <span class="connection-note">{{ selectedDatabase?.connection_name }} · live metadata · read-only</span>
      </div>

      <section v-if="tables.length" class="table-list">
        <article v-for="table in tables" :key="`${table.schema}.${table.name}`" class="table-card">
          <div class="table-heading">
            <div><p class="schema">{{ table.schema }}</p><h2>{{ table.name }}</h2></div>
            <span class="column-count">{{ table.columns.length }} columns</span>
          </div>
          <p v-if="table.description" class="description">{{ table.description }}</p>
          <div class="columns">
            <div class="column-row column-header"><span>Column</span><span>Type</span><span>Nullable</span><span>Description</span></div>
            <div v-for="column in table.columns" :key="column.name" class="column-row">
              <span class="column-name">{{ column.name }} <span v-if="column.primary_key" class="badge">PK</span></span>
              <code>{{ column.type }}</code><span>{{ column.nullable ? 'yes' : 'no' }}</span><span class="muted">{{ column.description || 'No description yet' }}</span>
            </div>
          </div>
          <p v-if="table.relations?.length" class="relations"><strong>Relations</strong><span v-for="relation in table.relations" :key="`${relation.from_column}-${relation.to_table}`">{{ relation.from_column }} → {{ relation.to_table }}.{{ relation.to_column }}</span></p>
        </article>
      </section>
      <p v-else class="empty">No tables found in the configured schemas.</p>
    </template>
  </main>
</template>