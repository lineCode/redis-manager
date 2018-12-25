<template>
  <div class="app-container">
    <el-card class="box-card container" style="position: relative;">
      <div slot="header" class="clearfix">
        <el-row :gutter="32">
          <el-col :xs="24" :sm="24" :lg="24" style="margin: 0 0 10px 0">
            <el-button style="width: 100%" class="modify-btn" size="mini" type="success" icon="el-icon-plus" @click="dialogFormVisible=true">Add row</el-button>
          </el-col>
          <el-col :xs="24" :sm="24" :lg="24" style="margin: 0 0 10px 0">
            <el-input v-model="keyword" prefix-icon="el-icon-search" size="mini" placeholder="请输入内容" />
          </el-col>
          <el-col :xs="24" :sm="24" :lg="24">
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleScanKeys"/>
            <el-button type="danger" icon="el-icon-delete" size="mini" @click="handleDelKeys"/>
            <el-button type="warning" size="mini" @click="handleLoadKeys">More</el-button>
          </el-col>
        </el-row>
      </div>
      <el-checkbox :indeterminate="isIndeterminate" v-model="checkAll" @change="handleCheckAllChange">全选 ({{ checkedKeysNum }}/{{ currentKeysNum }}/{{ dbSize }})</el-checkbox>
      <el-scrollbar wrap-style=" height: 74vh;">
        <el-checkbox-group v-model="checkedKeys" @change="handleCheckedChange">
          <ul style="padding: 10px 0;">
            <li>
              <el-checkbox v-for="key in keys" :label="key" :key="key" style="display: block; margin-left: 0;">
                <span @click.stop.prevent="keyClick(key)">
                  {{ key }}
                </span>
              </el-checkbox>
            </li>
          </ul>
        </el-checkbox-group>
      </el-scrollbar>
    </el-card>

    <el-dialog :visible.sync="dialogFormVisible" title="Add new key">
      <el-form ref="addForm" :model="addForm" size="small">
        <el-form-item label="key:">
          <el-input
            v-model="addForm.key"
            :rules="[
              { required: true, message: 'key不能为空'}
            ]"
            autocomplete="off"/>
        </el-form-item>
        <el-form-item label="type:" >
          <el-select v-model="addForm.type">
            <el-option label="string" value="string"/>
            <el-option label="list" value="list"/>
            <el-option label="hash" value="hash"/>
            <el-option label="set" value="set"/>
            <el-option label="zset" value="zset"/>
          </el-select>
        </el-form-item>
        <el-form-item v-if="addForm.type === 'hash' || addForm.type === 'zset'" label="field:">
          <el-input
            v-model="addForm.field"
            :rules="[
              { required: (addForm.type === 'hash' || addForm.type === 'zset'), message: 'field不能为空'}
            ]"
            autocomplete="off"/>
        </el-form-item>
        <el-form-item v-if="addForm.type === 'zset'" label="score:">
          <el-input
            v-model="addForm.score"
            :rules="[
              { required: (addForm.type === 'zset'), message: 'score不能为空'},
              { type: 'number', message: 'score必须为数字值'}
            ]"
            autocomplete="off"/>
        </el-form-item>
        <el-form-item v-if="addForm.type !== 'zset'" label="val:">
          <el-input
            :rows="2"
            v-model="addForm.val"
            type="textarea"
            placeholder="请输入内容"/>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogFormVisible = false">取 消</el-button>
        <el-button type="primary" @click="handleAddKey()">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import path from 'path'
import { getDbSize } from '@/api/server'
import { addKey, scanKeys, delBackground } from '@/api/key'
import { unique } from '@/utils/index'

export default {
  name: 'KeyList',
  data() {
    return {
      loading: null,
      dbSize: 0,
      keyword: '',
      iter: false,
      checkAll: false,
      checkedKeys: [],
      keys: [],
      isIndeterminate: true,
      addForm: { type: 'string', val: '' },
      dialogFormVisible: false,
      serverID: null
    }
  },
  computed: {
    checkedKeysNum() {
      return this.checkedKeys ? this.checkedKeys.length : 0
    },
    currentKeysNum() {
      return this.keys ? this.keys.length : 0
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        const serverID = this.$route.params.serverID
        if (serverID !== this.serverID) {
          this.serverID = serverID
          this.resetIter()
          this.handleScanKeys()
          this.refreshDbSize()
        }
      }
    }
  },
  created() {
    this.serverID = this.$route.params.serverID
    this.resetIter()
    this.handleScanKeys()
    const tmp = () => {
      this.refreshDbSize()
      return tmp
    }
    this.interval = setInterval(tmp(), 10000)
  },
  methods: {
    resolvePath(key) {
      const serverID = this.$route.params.serverID
      return path.resolve(this.basePath, '/server/' + serverID + '/key' + key)
    },
    refreshDbSize() {
      let serverID = this.$route.params.serverID
      if (!serverID) {
        serverID = 0
      }
      getDbSize(serverID).then(response => {
        this.dbSize = response.data
      }).catch(error => {
        console.log(error)
      })
    },
    keyClick(key) {
      const serverID = this.$route.params.serverID
      this.$router.push({ name: 'key', params: { serverID: serverID, key: key }})
    },
    resetIter() {
      this.iter = 0
      this.keys = []
      this.checkedKeys = []
    },
    handleScanKeys() {
      this.resetIter()
      this.handleLoadKeys()
    },
    handleLoadKeys() {
      const params = {}
      params.keyword = this.keyword
      params.iter = this.iter
      const serverID = this.$route.params.serverID
      const loading = this.$loading({
        lock: true,
        text: 'Loading',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      })
      scanKeys(serverID, params).then((response) => {
        const data = response.data
        this.iter = data.cursor
        data.keys = data.keys.concat(this.keys)
        this.keys = unique(data.keys)
        loading.close()
      }).catch(() => {
        loading.close()
      })
    },
    handleDelKeys() {
      const serverID = this.$route.params.serverID
      const keys = this.checkedKeys
      delBackground(serverID, keys).then((response) => {
        this.$message({
          message: 'success, 删除操作将在后台运行!',
          type: 'success',
          duration: 1000
        })
      })
    },
    handleCheckAllChange(val) {
      this.checkedKeys = val ? this.keys : []
      this.isIndeterminate = false
    },
    handleCheckedChange(value) {
      const checkedCount = value.length
      this.checkAll = checkedCount === this.keys.length
      this.isIndeterminate = checkedCount > 0 && checkedCount < this.keys.length
    },
    handleAddKey() {
      this.$refs.addForm.validate(valid => {
        if (valid) {
          const serverID = this.$route.params.serverID
          const type = this.addForm.type
          const params = {}
          params.type = type
          switch (type) {
            case 'string':
              params.data = this.addForm.val
              break
            case 'list':
            case 'set':
              params.data = []
              params.data.push(this.addForm.val)
              break
            case 'zset':
              params.data = []
              params.data.push({ Score: parseInt(this.addForm.score), Member: this.addForm.field })
              break
            case 'hash':
              params.data = {}
              params.data[this.addForm.field] = this.addForm.val
          }
          params.key = this.addForm.key
          addKey(serverID, params).then((response) => {
            this.dialogFormVisible = false
            let isExists = false
            for (let i = 0; i < this.keys.length; i++) {
              if (this.keys[i] === params.key) {
                isExists = true
                break
              }
            }
            if (!isExists) {
              this.keys.unshift(params.key)
            }
            this.$message({
              message: 'add success',
              type: 'success',
              duration: 1000
            })
          })
        }
      })
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
.container{
  box-shadow: 0 2px 6px rgba(0,0,0,.2);
  border-color: #eee;
  height: 100%;
}
.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}
.clearfix:after {
  clear: both
}
</style>
