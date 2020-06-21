<template>
  <div class="container">
    <div class="row">
      <div class="col-xl">
        <h1 class="display-3" id="title">Server Checker</h1>
      </div>
    </div>
    <div class="row">
      <div class="col-xl">
        <button
          @click="peticion"
          :disabled="disableButtons"
          type="button"
          class="btn btn-dark listButtons"
        >Fetch checked pages</button>
        <button
          @click="toggleList"
          :disabled="disableButtons"
          type="button"
          class="btn btn-dark listButtons"
        >Toggle List</button>
        <div v-if="showList">
          <MyList :propList="list"></MyList>
        </div>
        <h1 class="display-4" id="title">Search for any website</h1>
        <form @submit.prevent="onSubmit">
          <div class="form-group">
            <input required="required" v-model="url" type="text" class="form-control" />
          </div>
          <button type="submit" :disabled="disableButtons" class="btn btn-dark botonesTabla">Submit</button>
          <button
            @click="toggleTable"
            :disabled="disableButtons"
            type="button"
            class="btn btn-dark botonesTabla"
          >Toggle Table</button>
        </form>
        <div v-if="showTable">
          <MyTable
            :propTableIP="tableIP"
            :propTableGrade="tableGrade"
            :propServersChanged="servers_changed"
            :propSsl="ssl_grade"
            :propPrevSsl="previous_ssl_grade"
            :propLogo="logo"
            :propTitle="title"
            :propDown="is_down"
            :propMessage="message"
          ></MyTable>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import MyList from "./MyList.vue";
import MyTable from "./MyTable.vue";

export default {
  name: "MyComponent",
  components: {
    MyList,
    MyTable
  },
  data() {
    return {
      url: "",
      list: [],
      tableIP: [],
      tableGrade: [],
      showList: false,
      showTable: false,
      servers_changed: "",
      ssl_grade: "",
      previous_ssl_grade: "",
      logo: "",
      title: "",
      is_down: "",
      message: "",
      disableButtons: false
    };
  },
  methods: {
    toggleList: function() {
      if (this.showList) {
        this.showList = false;
      } else {
        this.showList = true;
      }
    },
    toggleTable: function() {
      if (this.showTable) {
        this.showTable = false;
      } else {
        this.showTable = true;
      }
    },
    peticion: function() {
      this.disableButtons = true;
      const myRequest = new Request("http://localhost:8085/consultas", {
        method: "GET"
      });
      fetch(myRequest)
        .then(response => response.json())
        .then(json => {
          var arrayAnswer = [];
          json.items.forEach(element => arrayAnswer.push(element.name));
          this.list = arrayAnswer;
          this.showList = true;
          this.disableButtons = false;
        });
    },
    onSubmit: function() {
      this.disableButtons = true;
      var stringArreglado = this.url.slice(0);
      stringArreglado = stringArreglado.replace("www.", '');
      stringArreglado = stringArreglado.replace("http://", '');
      stringArreglado = stringArreglado.replace("https://", '');
      const myRequest = new Request(
        "http://localhost:8085/dominio/" + stringArreglado,
        {
          method: "GET"
        }
      );
      fetch(myRequest)
        .then(response => response.json())
        .then(json => {
          this.tableIP = [];
          this.tableGrade = [];
          this.servers_changed = "";
          this.ssl_grade = "";
          this.previous_ssl_grade = "";
          this.logo = "";
          this.title = "";
          this.is_down = false;
          this.message = "";

          if (json.message !== undefined) {
            this.message = json.message;
            this.showTable = true;
          } else {
            var arrayIP = [];
            var arrayRate = [];
            json.servers.forEach(element => {
              arrayIP.push(element.address);
              arrayRate.push(element.ssl_grade);
            });
            this.servers_changed = json.servers_changed;
            this.ssl_grade = json.ssl_grade;
            this.previous_ssl_grade = json.previous_ssl_grade;
            this.logo = json.logo;
            this.title = json.title;
            this.is_down = json.is_down;
            this.tableIP = arrayIP;
            this.tableGrade = arrayRate;
            this.showTable = true;
          }
          this.disableButtons = false;
        });
    }
  }
};
</script>

<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}

#title {
  margin-bottom: 1em;
}

.listButtons {
  margin-left: 1em;
  margin-right: 1em;
  margin-bottom: 2em;
}

.botonesTabla {
  margin: 1em;
}
</style>
