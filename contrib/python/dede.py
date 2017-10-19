import urllib2
import time
import json
from selenium import webdriver
from selenium.webdriver import ActionChains
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
import sys


class DedeChapterManager:

    def __init__(self, dede, chapterID):
        self.dede = dede
        self.chapterID = chapterID

    def __enter__(self):
        self.prevChapterID = self.dede.chapterID
        self.dede.chapterID = self.chapterID

    def __exit__(self, type, value, traceback):
        self.dede.chapterID = self.prevChapterID


class DedeSectionManager:

    def __init__(self, dede, sectionID):
        self.dede = dede
        self.sectionID = sectionID

    def __enter__(self):
        self.prevSectionID = self.dede.sectionID
        self.dede.sectionID = self.sectionID

    def __exit__(self, type, value, traceback):
        self.dede.sectionID = self.prevSectionID


class Dede:

    def __init__(self, endpoint, driver, sessionID):
        self.endpoint = endpoint
        self.driver = driver
        self.sessionID = sessionID
        self.chapterID = ''
        self.sectionID = ''

    def fake_mouse(self):
        return DedeFakeMouse(self)

    def terminal_manager(self):
        return DedeTerminalManager(self)

    def video_recorder(self):
        return DedeVideoRecorder(self)

    def chapter(self, chapterID):
        return DedeChapterManager(self, chapterID)

    def section(self, sectionID):
        return DedeSectionManager(self, sectionID)


class DedeFakeMouse:

    def __init__(self, dede):
        self.dede = dede

    def install(self):
        # TODO catch error
        print("%s/fake-mouse/install" % self.dede.endpoint)
        script = urllib2.urlopen(
            "%s/fake-mouse/install" % self.dede.endpoint).read()
        self.dede.driver.execute_script(script)

    def _fake_mouse_click_on(self, el):
        self.dede.driver.execute_async_script(
            "DedeFakeMouse.clickOn(arguments[0], arguments[1])", el)

    def _fake_mouse_move_on(self, el):
        self.dede.driver.execute_async_script(
            "DedeFakeMouse.moveOn(arguments[0], arguments[1])", el)

    def click_on(self, el):
        self._fake_mouse_click_on(el)
        el.click()

    def double_click_on(self, el):
        self._fake_mouse_click_on(el)
        el.double_click()

    def move_on(self, el):
        self._fake_mouse_move_on(el)


class DedeTerminalManagerTab:

    def __init__(self, dede, window_handle):
        self.dede = dede
        self.window_handle = window_handle

    def focus(self):
        self.dede.driver.switch_to_window(self.window_handle)

    def start_record(self):
        self.dede.driver.execute_script(
            "DedeTerminal.startRecord(%d, %d, %d)" %
            (self.dede.sessionID, self.dede.chapterID, self.dede.sectionID))

    def stop_record(self):
        self.dede.driver.execute_script("DedeTerminal.stopRecord()")

    def type(self, str):
        self.dede.driver.execute_async_script(
            "DedeTerminal.type(arguments[0], arguments[1])", str)

    def type_cmd(self, str):
        self.dede.driver.execute_async_script(
            "DedeTerminal.typeCmd(arguments[0], arguments[1])", str)

    def type_cmd_wait(self, str, regex):
        self.dede.driver.execute_async_script(
            "DedeTerminal.typeCmdWait(\
             arguments[0], arguments[1], arguments[2])", str, regex)


class DedeTerminalManager:

    def __init__(self, dede):
        self.dede = dede
        self.termIndex = 1

    def open_terminal_tab(
            self, title, width=1400, cols=2000, rows=40, delay=70):
        self.dede.driver.execute_script(
            "window.open('%s/terminal/%s?"
            "title=%s&width=%d&cols=%d&rows=%d&delay=%d')" %
            (self.dede.endpoint, self.termIndex,
             title, width, cols, rows, delay))
        self.termIndex += 1

        window_handle = self.dede.driver.window_handles[-1]
        tab = DedeTerminalManagerTab(self.dede, window_handle)
        self.dede.driver.switch_to_window(window_handle)

        return tab


class DedeVideoRecord:

    def __init__(self, dede):
        self.dede = dede

    def stop(self):
        # TODO catch error
        urllib2.urlopen(
            "%s/video/stop-record?sessionID=%s&chapterID=%s&sectionID=%s" %
            (self.dede.endpoint, self.dede.sessionID,
             self.dede.chapterID, self.dede.sectionID))


class DedeVideoRecorder:

    def __init__(self, dede):
        self.dede = dede

    def start_record(self):
        # TODO catch error
        urllib2.urlopen(
            "%s/video/start-record?sessionID=%s&chapterID=%s&sectionID=%s" %
            (self.dede.endpoint, self.dede.sessionID,
             self.dede.chapterID, self.dede.sectionID))
        return DedeVideoRecord(self.dede)


class SkydiveSelenium:

    def __init__(self, driver, client, fake_mouse):
        self.driver = driver
        self.client = client
        self.fake_mouse = fake_mouse

    def click_on_node_by_id(self, id, retry=10):
        el = driver.find_element_by_id("node-img-%s" % id)

        for i in range(0, retry):
            try:
                self.fake_mouse.click_on(el)
                return
            except:
                pass

    def click_on_node_by_gremlin(self, gremlin):
        self.click_on_node_by_id(self.client.get_node_id(gremlin))

    def expand_group_by_id(self, id):
        expanded = False
        self.click_on_node_by_id(id)
        while not expanded:
            try:
                el = driver.find_element_by_id("node-img-%s" % id)
                chain = ActionChains(self.driver)
                chain.key_down(Keys.ALT)
                chain.move_to_element(el)
                chain.click(el)
                chain.key_up(Keys.ALT)
                chain.perform()
                expanded = True
            except:
                pass
        self.click_on_node_by_id(id)

    def expand_group_by_gremlin(self, gremlin):
        self.expand_group_by_id(self.client.get_node_id(gremlin))

    def pin_node_by_id(self, id):
        pin = False
        self.click_on_node_by_id(id)
        while not pin:
            try:
                el = driver.find_element_by_id("node-img-%s" % id)
                chain = ActionChains(self.driver)
                chain.key_down(Keys.SHIFT)
                chain.move_to_element(el)
                chain.click(el)
                chain.key_up(Keys.SHIFT)
                chain.perform()
                pin = True
            except:
                pass
        self.click_on_node_by_id(id)

    def pin_node_by_gremlin(self, gremlin):
        self.pin_node_by_id(self.client.get_node_id(gremlin))

    def scroll_down_right_panel(self):
        self.driver.execute_script(
            "$('#right-panel').animate(\
            {scrollTop: $('#right-panel').get(0).scrollHeight}, 500);")

    def scroll_up_right_panel(self):
        self.driver.execute_script(
            "$('#right-panel').animate({scrollTop: 0}, -500);")

    def wait_for_element_by_id(self, id, retry=10, delay=0.1):
        for i in range(0, retry):
            try:
                return self.driver.find_element_by_id(id)
            except:
                time.sleep(delay)

    def select_node(self, id, gremlin):
        self.fake_mouse.click_on(
            self.wait_for_element_by_id(id))
        self.click_on_node_by_gremlin(gremlin)

    def fill_textbox_by_id(self, id, text):
        tb = self.driver.find_element_by_id(id)
        tb.clear()
        tb.send_keys(text)

    def create_capture(self, gremlin1, gremlin2="", bpf=""):
        try:
            self.driver.find_element_by_id("start-capture")
        except:
            self.fake_mouse.click_on(
                self.driver.find_element_by_id("create-capture"))

        self.select_node("node-selector-1", gremlin1)

        if gremlin2:
            self.select_node("node-selector-2", gremlin2)

        if bpf:
            self.fake_mouse.click_on(
                self.driver.find_element_by_id("capture-bpf"))
            self.fill_textbox_by_id("capture-bpf", bpf)

        self.fake_mouse.click_on(
            self.driver.find_element_by_id("start-capture"))

    def get_flow_row_ids(self, gremlin, retry=10, delay=0.5):
        for i in range(0, retry):
            uuids = self.client.get_flow_uuids(gremlin)
            if not uuids:
                time.sleep(delay)
                continue

            return ["flow-%s" % uuid for uuid in uuids]

    def expand_flow_row(self, gremlin, retry=10, delay=0.5):
        ids = self.get_flow_row_ids(gremlin, retry, delay)
        if not ids:
            return
        for id in ids:
            for i in range(0, retry):
                try:
                    self.fake_mouse.click_on(
                        self.driver.find_element_by_id(id)
                    )
                    return
                except:
                    pass


class SkydiveClient:

    def __init__(self, endpoint):
        self.endpoint = endpoint

    def get_node_id(self, gremlin):
        data = json.dumps(
            {"GremlinQuery": gremlin}
        )
        req = urllib2.Request("http://%s/api/topology" % self.endpoint,
                              data, {'Content-Type': 'application/json'})
        resp = urllib2.urlopen(req)
        data = json.load(resp)
        if not data:
            return

        return data[0]["ID"]

    def get_flow_uuids(self, gremlin):
        data = json.dumps(
            {"GremlinQuery": gremlin}
        )
        req = urllib2.Request("http://%s/api/topology" % self.endpoint,
                              data, {'Content-Type': 'application/json'})
        resp = urllib2.urlopen(req)
        data = json.load(resp)
        if not data:
            return

        return [flow["UUID"] for flow in data]

    def wait_for_node_id(self, gremlin):
        id = skydive_cli.get_node_id(gremlin)
        while not id:
            id = skydive_cli.get_node_id(gremlin)

if __name__ == '__main__':

    #driver = webdriver.Remote(
    #  command_executor='http://127.0.0.1:4444/wd/hub',
    #  desired_capabilities={"browserName": "chrome"})
    driver = webdriver.Chrome()

    time.sleep(2)

    #driver.maximize_window()
    driver.get("http://192.168.50.10:8082")
    driver.set_script_timeout(20)

    window_handle = driver.window_handles[-1]

    time.sleep(2)

    # install fake mouse on the WebUI
    dede = Dede("http://localhost:55555", driver, 1)
    fake_mouse = dede.fake_mouse()
    fake_mouse.install()

    skydive_cli = SkydiveClient("192.168.50.10:8082")
    skydive_sel = SkydiveSelenium(driver, skydive_cli, fake_mouse)

    time.sleep(2)

    # expand
    fake_mouse.click_on(driver.find_element_by_id('expand'))

    # zoom-fit
    fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

    time.sleep(5)

    skydive_sel.expand_group_by_gremlin(
        "G.V().Has('Name', Regex('mysql.*')).In().Has('Type', 'netns')"
    )
    time.sleep(2)

    # zoom-fit
    fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

    skydive_sel.click_on_node_by_gremlin(
        "G.V().Has('Name', Regex('mysql.*'))")
    time.sleep(2)

    skydive_sel.expand_group_by_gremlin(
        "G.V().Has('Name', Regex('wordpress.*')).In().Has('Type', 'netns')"
    )
    time.sleep(2)

    skydive_sel.click_on_node_by_gremlin(
        "G.V().Has('Name', Regex('wordpress.*'))")
    time.sleep(2)

    tab1 = dede.terminal_manager().open_terminal_tab('agent1')
    tab1.type_cmd_wait("ssh agent1", "vagrant")
    tab1.type_cmd_wait("docker network list", "vagrant")
    time.sleep(2)

    # expand the
    driver.switch_to_window(window_handle)
    skydive_sel.expand_group_by_gremlin(
        "G.V().Has('Name', Regex('mysql.*')).In().Out()."
        "Has('Name', 'eth2').Both().Has('Type', 'veth')."
        "In().Has('Type', 'netns')"
    )

    # zoom-fit
    fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

    skydive_sel.create_capture(
        "G.V().Has('Name', Regex('mysql.*')).In().Out().Has('Name', 'eth2')",
        "G.V().Has('Name', Regex('wordpress.*')).In().Out()."
        "Has('Name', 'eth2')",
        "port 3306")
    time.sleep(2)

    driver.execute_script(
        "window.open('http://192.168.50.20:7070')")
    time.sleep(2)

    driver.switch_to_window(window_handle)
    time.sleep(1)

    skydive_sel.scroll_down_right_panel()
    fake_mouse.click_on(driver.find_element_by_xpath(
        "//i[@class='fa fa-refresh']"))

    skydive_sel.expand_flow_row("G.Flows().Has('Transport', '3306')")
    skydive_sel.scroll_down_right_panel()
    time.sleep(1)

    fake_mouse.click_on(driver.find_element_by_xpath(
        "//span[@class='object-key' and text()='Transport']"))
    time.sleep(1)

    skydive_sel.scroll_up_right_panel()

    skydive_sel.create_capture(
        "G.V().Has('Name', Regex('mysql.*')).In().Out().Has('Name', 'eth2')."
        "Both().Has('Type', 'veth').In().Has('Type', 'netns').Out()."
        "Has('Type', 'vxlan')",
        "",
        "port 3306")
    time.sleep(1)

    tab1.focus()
    tab1.type_cmd_wait(
        "docker service create --name wordpress2 --network swarmnet "
        "--constraint \"node.hostname==agent2\" --publish 7071:80  "
        "-e WORDPRESS_DB_HOST=mysql "
        "-e WORDPRESS_DB_PASSWORD=password wordpress", "vagrant")
    time.sleep(1)

    driver.execute_script(
        "window.open('http://192.168.50.20:7071')")
    time.sleep(2)

    driver.switch_to_window(window_handle)

    sys.exit(0)

    with dede.chapter(1):
        #record = dede.video_recorder().start_record()

        # start swarm mysql + first wordpress
        tab1 = dede.terminal_manager().open_terminal_tab('agent1')
        tab1.type_cmd_wait("ssh agent1", "vagrant")

        tab2 = dede.terminal_manager().open_terminal_tab('swarmjoin')
        with dede.section(4):
            tab2.focus()
            tab2.type_cmd(
                "ssh agent1 \"docker swarm init --listen-addr 192.168.50.20 "
                "--advertise-addr 192.168.50.20\"")
            time.sleep(1)
            tab2.type_cmd(
                "token=$( ssh agent1 'docker swarm join-token -q worker' )")
            time.sleep(1)
            tab2.type_cmd(
                "ssh agent2 \"docker swarm join "
                "--token $token 192.168.50.20:2377\"")
            time.sleep(1)

            driver.switch_to_window(window_handle)
            time.sleep(2)

        with dede.section(5):
            tab1.focus()
            tab1.type_cmd_wait(
                "docker network create -d overlay swarmnet", "vagrant")
            time.sleep(1)

            driver.switch_to_window(window_handle)
            time.sleep(2)

            tab1.focus()
            tab1.type_cmd_wait(
                "docker service create --name mysql --network swarmnet "
                "--constraint \"node.hostname==agent1\" --publish 3306:3306 "
                "--env=\"MYSQL_ROOT_PASSWORD=password\" mysql", "vagrant")
            time.sleep(2)

            driver.switch_to_window(window_handle)
            time.sleep(2)

            skydive_sel.expand_group_by_gremlin(
                "G.V().Has('Name', Regex('mysql.*')).In().Has('Type', 'netns')"
            )
            time.sleep(2)

            skydive_sel.click_on_node_by_gremlin(
                "G.V().Has('Name', Regex('mysql.*'))")
            time.sleep(2)

            tab1.focus()
            tab1.type_cmd_wait(
                "docker service create --name wordpress1 --network swarmnet "
                "--constraint \"node.hostname==agent1\" --publish 7070:80  "
                "-e WORDPRESS_DB_HOST=mysql "
                "-e WORDPRESS_DB_PASSWORD=password wordpress", "vagrant")
            time.sleep(1)

        sys.exit(0)

        # Intro
        with dede.section(1):

            # expand
            fake_mouse.click_on(driver.find_element_by_id('expand'))

            # zoom-fit
            fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

            time.sleep(5)

            """
            # pin agent1/eth1
            skydive_sel.pin_node_by_gremlin("G.V().Has('Name', 'agent1')")

            # pin agent1/eth1
            skydive_sel.pin_node_by_gremlin(
                "G.V().Has('Name', 'agent1').Out().Has('Name', 'eth1')")

            # pin agent2
            skydive_sel.pin_node_by_gremlin("G.V().Has('Name', 'agent2')")

            # select agent2/eth1
            skydive_sel.pin_node_by_gremlin(
                "G.V().Has('Name', 'agent2').Out().Has('Name', 'eth1')")
            """

            # zoom-fit
            fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

            # click to be sure
            skydive_sel.click_on_node_by_gremlin(
                "G.V().Has('Name', 'agent2').Out().Has('Name', 'eth1')")

            # select metadata
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//h1[text()='metadatas']"))

            # show ipv4
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@class='object-key-value ipv4']\
                /div/span[@class='object-key']"))

            time.sleep(0.2)

            # show mtu
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@class='object-key-value mtu']\
                /div/span[@class='object-key']"))

            # show metrics
            skydive_sel.scroll_down_right_panel()

            time.sleep(0.2)

            # click on Fields
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@id='last-interface-metrics']\
                //button[@class='btn btn-default dropdown-toggle btn-xs']"))
            time.sleep(1)
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@id='last-interface-metrics']\
                //button[@class='btn btn-default dropdown-toggle btn-xs']"))

            skydive_sel.scroll_up_right_panel()

            time.sleep(1)

        tab1 = dede.terminal_manager().open_terminal_tab('agent1')

        # create tap interface
        with dede.section(2):
            tab1.focus()
            tab1.type_cmd_wait("ssh agent1", "vagrant")
            time.sleep(1)
            tab1.type_cmd_wait(
                "sudo ip tuntap add dev tap-demo mode tap", "vagrant")
            time.sleep(1)

            driver.switch_to_window(window_handle)
            time.sleep(1)

            # zoom-fit
            fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

            skydive_sel.click_on_node_by_gremlin(
                "G.V().Has('Name', 'tap-demo')")
            time.sleep(1)

            fake_mouse.click_on(driver.find_element_by_xpath(
                ".//h1[text()='metadatas']"))
            time.sleep(1)

            tab1.focus()
            tab1.type_cmd_wait("sudo ip link set tap-demo up", "vagrant")
            tab1.type_cmd_wait(
                "sudo ip addr add 10.0.0.99/32 dev tap-demo", "vagrant")
            time.sleep(1)

            driver.switch_to_window(window_handle)
            time.sleep(1)
            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@class='object-key-value ipv4']"
                "/div/span[@class='object-key']"))
            time.sleep(1)

            tab1.focus()
            tab1.type_cmd_wait("sudo ip link del tap-demo", "vagrant")
            time.sleep(0.5)

            driver.switch_to_window(window_handle)
            time.sleep(2)

        # start busybox
        with dede.section(3):
            tab1.focus()
            tab1.type_cmd_wait(
                "docker run --name busybox -d -it busybox", "vagrant")
            time.sleep(1)

            driver.switch_to_window(window_handle)

            skydive_cli.wait_for_node_id("G.V().Has('Type', 'netns')")

            # zoom-fit
            fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

            skydive_sel.click_on_node_by_id(id)
            time.sleep(1)

            skydive_sel.expand_group_by_gremlin("G.V().Has('Type', 'netns')")

            # zoom-fit
            fake_mouse.click_on(driver.find_element_by_id('zoom-fit'))

            time.sleep(2)

            skydive_sel.click_on_node_by_gremlin(
                "G.V().Has('Name', 'busybox')")

            time.sleep(2)

            fake_mouse.click_on(driver.find_element_by_xpath(
                "//div[@class='object-key-value containername']\
                /div/span[@class='object-key']"))

        driver.switch_to_window(window_handle)

        sys.exit(0)

        time.sleep(20)

        # start 1st wordpress
        with dede.section(2):
            tab1.focus()
            tab1.type_cmd_wait("docker service create --name wordpress1 --network swarmnet --constraint 'node.hostname==agent1' --publish 7070:80  -e WORDPRESS_DB_HOST=mysql -e WORDPRESS_DB_PASSWORD=password wordpress", "vagrant")
            time.sleep(1)

        driver.switch_to_window(window_handle)

        time.sleep(20)

        # start 2nd wordpress
        with dede.section(2):
            tab1.focus()
            tab1.start_record()
            tab1.type_cmd_wait("docker service create --name wordpress2 --network swarmnet --constraint 'node.hostname==agent2' --publish 7071:80  -e WORDPRESS_DB_HOST=mysql -e WORDPRESS_DB_PASSWORD=password wordpress", "vagrant")
            time.sleep(1)
            tab1.stop_record()

        driver.switch_to_window(window_handle)

        time.sleep(20)

    #record.stop()

    time.sleep(10)

    driver.close()
