<template>
  <div class="app-container">
    <el-card
      v-loading="loading"
      v-if="info.type && info.type !== 'none'"
      class="box-card container"
      element-loading-text="loading"
      element-loading-spinner="el-icon-loading"
      element-loading-background="rgba(0, 0, 0, 0.8)">
      <div slot="header" class="clearfix" style="line-height: 30px;">
        <el-row :gutter="10" type="flex" class="row-bg" justify="end">
          <el-col :span="1">
            <span><b>{{ info.type }}:</b></span>
          </el-col>
          <el-col :span="7">
            <el-input v-model="info.key" size="small" />
          </el-col>
          <el-col :span="3">
            <span><b>Size: </b>{{ info.size }}</span>
          </el-col>
          <el-col :span="3">
            <span><b>ttl:</b>{{ info.ttl }}</span>
          </el-col>
          <el-col :span="8">
            <el-button size="small" type="primary" plain @click="handleRename">rename</el-button>
            <el-button size="small" type="info" plain @click="reload">reload</el-button>
            <el-button size="small" type="warning" plain @click="handleExpire">set ttl</el-button>
            <el-button size="small" type="danger" plain @click="handleDel">del</el-button>
          </el-col>
          <el-col :span="2" />
        </el-row>
      </div>
      <div style="text-align: left; width: 100%;height: 90%;">
        <el-row v-if="info.type !== 'string'" :gutter="10" style="margin: 10px 0; text-align: left;">
          <el-col :span="18">
            <el-table
              v-loading="scanInfo.loading"
              ref="scanValTable"
              :data="info.data"
              :height="300"
              element-loading-text="loading"
              element-loading-spinner="el-icon-loading"
              element-loading-background="rgba(0, 0, 0, 0.8)"
              stripe
              border
              size="mini"
              highlight-current-row
              style="width: 100%;"
              @current-change="handleCurrentChange"
            >
              <el-table-column
                prop="index"
                label="index"/>
              <el-table-column
                v-if="info.type === 'hash' || info.type === 'zset'"
                prop="field"
                label="field"/>
              <el-table-column
                v-if="info.type === 'zset'"
                prop="score"
                label="score"/>
              <el-table-column
                v-if="info.type !== 'zset'"
                prop="val"
                label="val"/>
              <el-table-column
                align="right">
                <template slot-scope="scope">
                  <el-button
                    size="mini"
                    type="danger"
                    @click="handleDeleteRow(scope.$index, scope.row)">Delete</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-col>
          <el-col :span="4" >
            <el-button size="small" class="modify-btn" type="primary" icon="el-icon-plus" @click="dialogFormVisible=true">Add row</el-button>
            <div v-if="info.type !== 'list'" style="margin: 10px 0;">
              <el-input v-model="scanInfo.keyword" size="small" placeholder="search"/>
            </div>
            <div v-if="info.type === 'zset'" style="margin: 10px 0;">
              Order：<el-checkbox v-model="orderAsc">Asc</el-checkbox>
            </div>
            <div style="margin: 10px 0;">
              <el-row v-if="info.type === 'zset' || info.type === 'list'">
                <el-col :span="6" >
                  range :
                </el-col>
                <el-col :span="18" >
                  <el-input-number v-model="scanInfo.offset" :controls="false" style="width: 80px;" label="page" size="small"/>
                  <el-input-number v-model="scanInfo.num" :controls="false" style="width: 80px;" label="page" size="small"/>
                </el-col>
              </el-row>
              <el-button size="small" class="modify-btn" type="warning" icon="el-icon-search" @click="handleReloadValue">Reload value</el-button>
              <el-button v-if="info.type === 'hash' || info.type === 'set'" size="small" class="modify-btn" type="info" icon="el-icon-more-outline" @click="handleLoadMore">Load More</el-button>
            </div>
          </el-col>
          <el-col :span="2" />
        </el-row>
        <el-row v-if="info.type === 'hash' || info.type === 'zset'" :gutter="10" style="margin: 10px 0; text-align: left;">
          <el-col :span="22">
            <div>Key: size in bytes: {{ fieldSize }}</div>
            <div>
              <el-input v-model="editInfo.field" size="small" placeholder="请输入内容"/>
            </div>
          </el-col>
          <el-col :span="2" />
        </el-row>
        <el-row v-if="info.type === 'zset'" :gutter="10" style="margin: 10px 0; text-align: left;">
          <el-col :span="22">
            <div>Score: </div>
            <div>
              <el-input v-model="editInfo.score" size="small" placeholder="请输入score"/>
            </div>
          </el-col>
          <el-col :span="2" />
        </el-row>
        <el-row v-if="info.type !== 'zset'" :gutter="10" style="margin: 10px 0; text-align: left;">
          <el-col :span="22">
            <div>Value: size in bytes: {{ valueSize }}</div>
            <div>
              <el-input
                :rows="15"
                v-model="editInfo.val"
                size="small"
                type="textarea"
                placeholder="请输入内容"/>
            </div>
          </el-col>
          <el-col :span="2" />
        </el-row>
        <el-row :gutter="10" style="margin: 10px 0; text-align: right;">
          <el-col :span="22">
            <el-button size="small" type="primary" plain @click="handleSaveRow()">save</el-button>
          </el-col>
          <el-col :span="2" />
        </el-row>
      </div>
      <el-dialog :visible.sync="dialogFormVisible" title="Add row">
        <el-form ref="addRowForm" :model="addForm" size="small">
          <el-form-item v-if="info.type === 'hash' || info.type === 'zset'" label="field:">
            <el-input
              v-model="addForm.field"
              :rules="[
                { required: (addForm.type === 'hash' || info.type === 'zset'), message: 'field不能为空'}
              ]"
              autocomplete="off"/>
          </el-form-item>
          <el-form-item v-if="info.type === 'zset'" label="score:">
            <el-input
              v-model="addForm.score"
              :rules="[
                { required: (info.type === 'zset'), message: 'score不能为空'},
                { type: 'number', message: 'score必须为数字值'}
              ]"
              autocomplete="off"/>
          </el-form-item>
          <el-form-item v-if="info.type !== 'zset'" label="val:">
            <el-input
              :rows="2"
              v-model="addForm.val"
              type="textarea"
              placeholder="请输入内容"/>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="small" @click="dialogFormVisible = false">取 消</el-button>
          <el-button size="small" type="primary" @click="handleAddRow()">确 定</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>
<script>
import splitPane from 'vue-splitpane'
import { saveKey, getKey, saveRow, renameKey, delKey, delRow, scanVal, setTTL } from '@/api/key'
import { sizeof } from '@/utils/index'

export default {
  name: 'Edit',
  components: { splitPane },
  data() {
    return {
      loading: false,
      orderAsc: false,
      dialogFormVisible: false,
      editInfo: { field: '', val: '', score: 0 },
      addForm: { field: '', val: '', score: 0 },
      info: { },
      hashVal: [],
      scanInfo: {
        loading: false,
        keyword: '',
        iter: 0,
        offset: 0,
        num: 100
      }
    }
  },
  computed: {
    fieldSize() {
      return this.editInfo.field ? parseInt(sizeof(this.editInfo.field)) : 0
    },
    valueSize() {
      return this.editInfo.val ? parseInt(sizeof(this.editInfo.val)) : 0
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        this.reload()
      }
    }
  },
  created() {
    this.reload()
  },
  methods: {
    reload() {
      this.scanInfo = {
        iter: 0,
        offset: 0,
        num: 100,
        keyword: ''
      }
      this.resetSeleteted()
      if (!this.$route.params.key) {
        return
      }
      this.loading = true
      this.handleGetKey().then(() => {
        this.loading = false
        this.handleReloadValue().then(() => {
        })
      }).catch(() => {
        this.loading = false
      })
    },
    handleReloadValue() {
      this.scanInfo.iter = 0
      this.info.data = []
      this.handleScanValue().then(() => {
      })
    },
    handleLoadMore() {
      if (this.info.type === 'hash' || this.info.type === 'set') {
        this.handleScanValue().then((data) => {
        })
      }
    },
    handleScanValue() {
      return new Promise((resolve, reject) => {
        if (this.info.type === 'none') {
          this.$message({
            type: 'error',
            message: 'the key [' + key + '] is not exits'
          })
          reject(new Error('the key [' + key + '] is not exits'))
          return
        }
        const serverID = this.$route.params.serverID
        const key = this.$route.params.key
        const params = {
          key,
          keyword: this.scanInfo.keyword,
          start: this.scanInfo.offset,
          iter: this.scanInfo.iter,
          asc: this.orderAsc,
          num: this.scanInfo.num
        }
        this.scanInfo.loading = true
        scanVal(serverID, params).then((response) => {
          const data = response.data
          this.scanInfo.iter = data.cursor
          const tmp = []
          switch (this.info.type) {
            case 'string':
              this.editInfo.val = data.val
              break
            case 'set':
              for (let i = 0; i < data.val.length; i++) {
                tmp.push({ val: data.val[i] })
              }
              break
            case 'zset':
              for (let i = 0; i < data.val.length; i++) {
                tmp.push({ field: data.val[i].Member, score: data.val[i].Score })
              }
              break
            case 'list':
              for (let i = 0; i < data.val.length; i++) {
                tmp.push({ val: data.val[i].val, index: data.val[i].index })
              }
              break
            case 'hash':
              for (const item in data.val) {
                tmp.push({ val: data.val[item], field: item })
              }
              break
          }
          const unique = (items) => {
            const result = {}
            const finalResult = []
            let index = 0
            for (let i = 0; i < items.length; i++) {
              switch (this.info.type) {
                case 'list':
                  result[items[i].index] = items[i]
                  break
                case 'set':
                  items[i].index = index++
                  result[items[i].val] = items[i]
                  break
                case 'zset':
                case 'hash':
                  items[i].index = index++
                  result[items[i].field] = items[i]
                  break
              }
            }
            for (const item in result) {
              finalResult.push(result[item])
            }
            return finalResult
          }
          this.info.data = unique(tmp.concat(this.info.data))
          this.scanInfo.loading = false
          resolve(response)
        }).catch(error => {
          this.scanInfo.loading = false
          reject(error)
        })
      })
    },
    handleGetKey() {
      return new Promise((resolve, reject) => {
        const serverID = this.$route.params.serverID
        const key = this.$route.params.key
        getKey(serverID, key).then((response) => {
          const data = response.data
          this.info = data
          this.editInfo.val = data.data
          if (data.type === 'none') {
            this.$message({
              type: 'error',
              message: 'the key [' + key + '] is not exits'
            })
            reject(new Error('the key [' + key + '] is not exits'))
          } else {
            resolve()
          }
        }).catch(error => {
          reject(error)
        })
      })
    },
    handleSaveRow() {
      const serverID = this.$route.params.serverID
      const params = {}
      params.key = this.info.key
      params.type = this.info.type
      const tmpRow = {}
      switch (this.info.type) {
        case 'string':
          params.data = this.editInfo.val
          break
        case 'list':
          params.data = { val: this.editInfo.val, index: this.editInfo.selectedIndex }
          params.field = this.editInfo.selectedField
          tmpRow.val = this.editInfo.val
          break
        case 'set':
          params.field = this.editInfo.selectedField
          params.data = []
          params.data.push(this.editInfo.val)
          tmpRow.val = this.editInfo.val
          break
        case 'zset':
          params.field = this.editInfo.selectedField
          params.data = []
          params.data.push({ Score: parseInt(this.editInfo.score), Member: this.editInfo.field })
          tmpRow.score = this.editInfo.score
          tmpRow.field = this.editInfo.field
          break
        case 'hash':
          params.field = this.editInfo.selectedField
          params.data = {}
          params.data[this.editInfo.field] = this.editInfo.val
          tmpRow.field = this.editInfo.field
          tmpRow.val = this.editInfo.val
          break
        case 'none':
          this.$message({
            type: 'error',
            message: 'the key is not exists'
          })
          return
      }
      saveRow(serverID, params).then((response) => {
        this.$message({
          type: 'success',
          message: 'save success'
        })
        this.editInfo.selectedRow.field = tmpRow.field
        this.editInfo.selectedRow.val = tmpRow.val
        this.editInfo.selectedRow.score = tmpRow.score
      })
    },
    handleAddRow() {
      const serverID = this.$route.params.serverID
      const params = {}
      params.key = this.info.key
      params.type = this.info.type
      let row = {}
      switch (this.info.type) {
        case 'string':
          params.data = this.addForm.val
          break
        case 'list':
        case 'set':
          params.data = []
          params.data.push(this.addForm.val)
          row.val = this.addForm.val
          break
        case 'zset':
          params.data = []
          params.data.push({ Score: parseInt(this.addForm.score), Member: this.addForm.field })
          row = { field: this.addForm.field, score: this.addForm.score }
          break
        case 'hash':
          params.data = {}
          params.data[this.addForm.field] = this.addForm.val
          row = { field: this.addForm.field, val: this.addForm.val }
          break
        case 'none':
          this.$message({
            type: 'error',
            message: 'the key is not exists'
          })
          return
      }
      this.scanInfo.loading = true
      saveKey(serverID, params).then((response) => {
        this.$message({
          type: 'success',
          message: 'add success'
        })
        this.dialogFormVisible = false
        this.scanInfo.loading = false
        this.info.size++
        if (this.info.type === 'list') {
          row.index = this.info.size - 1
        } else {
          row.index = this.info.data.length
        }
        this.info.data.unshift(row)
      }).catch(() => {
        this.dialogFormVisible = false
        this.scanInfo.loading = false
      })
    },
    handleRename() {
      this.$prompt('New key', 'Rename key', {
        confirmButtonText: 'submit',
        cancelButtonText: 'cancel',
        inputPattern: /.+/,
        inputErrorMessage: 'key can not be empty'
      }).then(({ value }) => {
        const serverID = this.$route.params.serverID
        renameKey(serverID, this.info.key, value).then((response) => {
          this.$message({
            type: 'success',
            message: 'success '
          })
          this.$router.push({ name: 'key', params: { serverID: serverID, key: value }})
        })
      })
    },
    handleExpire() {
      this.$prompt('TTL', 'Set ttl', {
        confirmButtonText: 'submit',
        cancelButtonText: 'cancel',
        inputPattern: /[-]?[1-9]+[0-9]*/,
        inputErrorMessage: 'ttl must be int'
      }).then(({ value }) => {
        const serverID = this.$route.params.serverID
        setTTL(serverID, this.info.key, value).then((response) => {
          this.$message({
            type: 'success',
            message: 'set ttl success '
          })
          this.info.ttl = value
        })
      })
    },
    handleDel() {
      this.$confirm('确定删除该key？', 'Confirm', {
        confirmButtonText: 'submit',
        cancelButtonText: 'cancel',
        type: 'warning'
      }).then(() => {
        const serverID = this.$route.params.serverID
        delKey(serverID, this.info.key).then((response) => {
          this.$message({
            type: 'success',
            message: 'del success '
          })
          this.reload()
        })
      })
    },
    resetSeleteted() {
      this.editInfo.selectedIndex = 0
      this.editInfo.selectedField = ''
      this.editInfo.val = ''
      this.editInfo.field = ''
      this.editInfo.score = 0
      this.editInfo.selectedRow = null
    },
    handleCurrentChange(row) {
      if (!row) {
        this.resetSeleteted()
        return
      }
      this.editInfo.selectedRow = row
      switch (this.info.type) {
        case 'string':
          break
        case 'list':
          this.editInfo.selectedIndex = row.index
          this.editInfo.val = row.val
          this.editInfo.selectedField = row.val
          break
        case 'set':
          this.editInfo.selectedField = row.val
          this.editInfo.val = row.val
          break
        case 'zset':
          this.editInfo.selectedField = row.field
          this.editInfo.field = row.field
          this.editInfo.score = row.score
          break
        case 'hash':
          this.editInfo.selectedField = row.field
          this.editInfo.field = row.field
          this.editInfo.val = row.val
          break
        case 'none':
          this.$message({
            type: 'error',
            message: 'the key is not exists'
          })
      }
    },
    handleDeleteRow(index, row) {
      const serverID = this.$route.params.serverID
      const params = {}
      params.key = this.info.key
      params.type = this.info.type
      switch (this.info.type) {
        case 'string':
          break
        case 'list':
          params.index = row.index
          params.field = row.val
          break
        case 'set':
          params.field = row.val
          break
        case 'hash':
        case 'zset':
          params.field = row.field
          break
        case 'none':
          this.$message({
            type: 'error',
            message: 'the key is not exists'
          })
          return
      }
      this.scanInfo.loading = true
      delRow(serverID, params).then(() => {
        this.info.data.splice(index, 1)
        this.scanInfo.loading = false
      }).catch(() => {
        this.scanInfo.loading = false
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
.components-container {
  position: relative;
  height: 100vh;
}

.left-container {
  height: 100%;
}

.right-container {
  height: 200px;
}

.top-container {
  width: 100%;
  height: 100%;
}

.bottom-container {
  width: 100%;
  height: 100%;
}
.left{
  float: left;
}
.modify-btn{
  display: block; margin: 10px 0;
  width: 100%;
}

.container:hover{
  box-shadow: 0 2px 6px rgba(0,0,0,.2);
  border-color: #eee;
}
</style>

