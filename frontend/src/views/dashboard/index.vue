<template>
  <div class="dashboard-container">
    <el-row :gutter="32">
      <el-col v-for="(item) in serverList" :key="item.id" :xs="24" :sm="24" :lg="8">
        <el-card class="box-card-component" style="margin-left:8px;">
          <div slot="header" class="box-card-header">
            <div style="border-bottom: 1px solid #ebeef5;">
              <span @click="resolvePath(item.id)">
                <mallki :text="item.name" class-name="mallki-text"/>
              </span>
              <span style="float: right;">
                <el-button v-if="item.preload" :disabled="item.refresh_status === 1" class="modify-btn" type="warning" size="small" @click="handleRefreshKeys(item)">
                  <i v-if="item.refresh_status === 1" class="el-icon-loading"/>
                  <i v-else class="el-icon-refresh"/>
                  refresh
                </el-button>
                <el-tag v-if="item.initialize === false" type="pending"> pending  </el-tag>
                <el-button class="modify-btn" type="primary" icon="el-icon-setting" size="small" @click="opentty(item.id)">tty</el-button>
              </span>
              <div style="clear: both;height: 0px;"/>
            </div>
            <div style="clear: both; padding: 10px 0 0 0;margin: 5px 0 0 0">
              <div v-if="item.last_refresh_time > 0" style="text-align: right">
                <span>刷新时间：</span>
                <span>{{ formatTime(item.last_refresh_time) }}</span>
              </div>
              <div class="progress-item">
                <span>Memory</span>
                <el-progress :percentage="Math.floor(item.memory_used_percent)"/>
              </div>
            </div>
          </div>
          <div>
            <div v-for="task in item.del_task" :key="task.id" class="progress-item">
              <span>{{ task.task_name }}</span>
              <el-row :gutter="32">
                <el-col :span="24">
                  <el-progress :percentage="Math.floor(task.process)"/>
                </el-col>
                <button v-if="task.status == 0" class="stop" @click="handleStopDelTask(item.id, task.task_id )"/>
                <button v-else class="destroy" @click="handleRemoveDelTask(item.id, task.task_id )"/>
              </el-row>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-dialog
      :visible.sync="infoDialog"
      title="info"
      width="90%"
      center>
      <el-row :gutter="32">
        <el-col v-for="(item, index) in info" :key="index" :xs="24" :sm="24" :lg="6">
          <el-card class="box-card-component" style="margin-left:8px;">
            <div slot="header" class="box-card-header">
              <mallki :text="index" class-name="mallki-text"/>
              <div style="clear: both;height: 0px;"/>
            </div>
            <div style="clear: both;">
              <div v-for="(detail, index) in item" :key="index">{{ detail }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { getServerInfo, getServerList, stopDelTask, removeDelTask, refreshKeys } from '@/api/server'
import Mallki from '@/components/TextHoverEffect/Mallki'
import { formatTime } from '@/utils/index'

export default {
  name: 'Dashboard',
  components: { Mallki },
  data() {
    return {
      interval: null,
      infoDialog: false,
      info: {},
      loading: false,
      serverList: []
    }
  },
  computed: {
    ...mapGetters([
      'name',
      'roles'
    ])
  },
  created() {
    const tmp = () => {
      getServerList().then(response => {
        this.serverList = response.data
      })
      return tmp
    }
    this.interval = setInterval(tmp(), 1000)
  },
  beforeDestroy() {
    clearInterval(this.interval)
  },
  methods: {
    handleStopDelTask(serverID, taskID) {
      stopDelTask(serverID, taskID).then(() => {
        this.$message({
          message: 'success',
          type: 'success',
          duration: 1000
        })
      })
    },
    handleRemoveDelTask(serverID, taskID) {
      removeDelTask(serverID, taskID).then(() => {
        this.$message({
          message: 'success',
          type: 'success',
          duration: 1000
        })
      })
    },
    handleShowServerInfo(serverID) {
      getServerInfo(serverID).then(response => {
        const data = response.data
        const tmp = data.split('\r\n')
        let section = ''
        for (let i = 0; i < tmp.length; i++) {
          if (tmp[i].indexOf('#') === 0) {
            section = tmp[i]
            this.info[section] = []
            continue
          }
          this.info[section].push(tmp[i])
        }
        this.infoDialog = true
      })
    },
    handleRefreshKeys(item) {
      refreshKeys(item.id).then(resposne => {
        this.$message({
          message: 'success, refresh running at backend!',
          type: 'success',
          duration: 1000
        })
        item.refresh_status = 1
      })
    },
    resolvePath(serverID) {
      this.$router.push({ name: 'main', params: { serverID: serverID }})
    },
    formatTime(time, cFormat) {
      return formatTime(time, cFormat)
    },
    opentty(serverID) {
      const left = (document.body.clientWidth - 900) / 2
      window.open('/tty/?server_id=' + serverID, '_blank',
        'toolbar=yes, location=yes, directories=no, status=no, menubar=yes, scrollbars=yes, resizable=no, copyhistory=yes, width=900, height=500, top=100, left=' + left)
    }
  }
}
</script>

<style rel="stylesheet/scss" lang="scss" scoped>
  .dashboard {
    &-container {
      margin: 30px;
    }
    &-text {
      font-size: 30px;
      line-height: 46px;
    }
  }
  .box-card-component{
    margin-bottom: 10px;
  }

  .panel-group {
    margin-top: 18px;
    .card-panel-col {
      margin-bottom: 32px;
    }
    .card-panel {
      height: 108px;
      cursor: pointer;
      font-size: 12px;
      position: relative;
      overflow: hidden;
      color: #666;
      background: #fff;
      box-shadow: 4px 4px 40px rgba(0, 0, 0, .05);
      border-color: rgba(0, 0, 0, .05);
      &:hover {
        .card-panel-icon-wrapper {
          color: #fff;
        }
        .icon-people {
          background: #40c9c6;
        }
        .icon-message {
          background: #36a3f7;
        }
        .icon-money {
          background: #f4516c;
        }
        .icon-shopping {
          background: #34bfa3
        }
      }
      .icon-people {
        color: #40c9c6;
      }
      .icon-message {
        color: #36a3f7;
      }
      .icon-money {
        color: #f4516c;
      }
      .icon-shopping {
        color: #34bfa3
      }
      .card-panel-icon-wrapper {
        float: left;
        margin: 14px 0 0 14px;
        padding: 16px;
        transition: all 0.38s ease-out;
        border-radius: 6px;
      }
      .card-panel-icon {
        float: left;
        font-size: 48px;
      }
      .card-panel-description {
        float: right;
        font-weight: bold;
        margin: 26px;
        margin-left: 0px;
        .card-panel-text {
          line-height: 18px;
          color: rgba(0, 0, 0, 0.45);
          font-size: 16px;
          margin-bottom: 12px;
        }
        .card-panel-num {
          font-size: 20px;
        }
      }
    }
  }

  .progress-item .destroy {
    display: none;
    position: absolute;
    top: 0;
    right: 0px;
    bottom: 0;
    width: 40px;
    height: 40px;
    margin: auto 0;
    font-size: 23px;
    color: #cc9a9a;
    transition: color 0.2s ease-out;
    cursor: pointer;
    border: none;
    background: none;
  }
  .progress-item .destroy:hover {
    color: #af5b5e;
  }
  .progress-item .destroy:after {
    content: '×';
  }
  .progress-item:hover .destroy {
    display: block;
  }

  .progress-item .stop {
    display: none;
    position: absolute;
    top: 0;
    right: 0px;
    bottom: 0;
    width: 40px;
    height: 40px;
    margin: auto 0;
    font-size: 20px;
    color: #cc9a9a;
    transition: color 0.2s ease-out;
    cursor: pointer;
    border: none;
    background: none;
  }
  .progress-item .stop:hover {
    color: #af5b5e;
  }
  .progress-item .stop:after {
    content: 'Θ';
  }
  .progress-item:hover .stop {
    display: block;
  }

</style>
