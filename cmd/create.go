/*
Copyright © 2021 Name HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"libvirt.org/libvirt-go"
)

var (
	Num      int
	base     string
	cpu      int
	mem      int
	macAddr  string
	image    string
	userData string
	metaData string
	domImage string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create VM (Virtual Machine)",
	Run: func(cmd *cobra.Command, args []string) {

		Datadir, NetAddr = GetCFG()

		if Num == 0 {
			if _, err := os.Stat(Datadir + "/images/" + image); os.IsNotExist(err) {
				fmt.Println("Don't Create 'VM' Only Create Image")
				fmt.Printf("'%s' is not exist. 'virt-go' attempd to create image via 'base' image file. \n Enter base image full path : ", image)
				fmt.Scanf("%s", &base)
				GenImage(base, image)
				os.Exit(0)
			} else {
				fmt.Printf("%s is already exists! \n", image)
				os.Exit(0)
			}
		}

		macAddr = GetMAC(Num)
		userData = Datadir + "/cloudinit/user-data"
		metaData = Datadir + "/cloudinit/meta-data"

		// about image
		if _, err := os.Stat(Datadir + "/images/" + image); os.IsNotExist(err) {
			//fmt.Println(Datadir + "/images/" + image)
			fmt.Printf("'%s' is not exist. 'virt-go' attempd to create image via 'base' image file. \n Enter base image full path : ", image)
			fmt.Scanf("%s", &base)
			GenImage(base, image)
			domImage = GenDomDisk(image, Num)
		} else {
			domImage = GenDomDisk(image, Num)
		}

		mf, _ := os.Create(metaData)
		mf.WriteString("local-hostname: " + "virt-go-" + image + "-" + strconv.Itoa(Num))
		mf.Close()

		// Get connection libvirt
		conn, err := libvirt.NewConnect("qemu:///system")
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()

		// Define ISO
		isoFile, err := GenISOXML(GenISO(Num, image, userData, metaData))
		if err != nil {
			fmt.Println(err)
		}

		// Define Domain
		dom, err := conn.DomainDefineXML(GenDomXML(image, Num, domImage, cpu, mem, macAddr))
		if err != nil {
			fmt.Println(err)
		}

		// Update Domain to use ISO File
		err = dom.UpdateDeviceFlags(isoFile, 0)
		if err != nil {
			fmt.Println(err)
		}

		// Start Domain
		err = dom.Create()
		if err != nil {
			fmt.Println(err)
		}

		// Detach ISO file for not using when reboot
		/*
		   	emptyCDrom := `<disk type="file" device="cdrom">
		     <source></source>
		     <driver Name="qemu" type="raw"></driver>
		     <backingStore/>
		     <target dev="hda" bus="ide"></target>
		     <readonly></readonly>
		     <address type="drive" controller="0" bus="0" target="0" unit="0"></address>
		   	</disk>`
		*/
		//err = dom.UpdateDeviceFlags(emptyCDrom, 0)
		//if err != nil {
		//	fmt.Println(err)
		//}

		// Print Result
		resultName, err := dom.GetName()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("\"%s\" is created! \n", resultName)

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	// Num      int
	// base     string
	// cpu      int
	// mem      int
	// macAddr  string
	// image    string
	// userData string
	// metaData string
	// domImage string
	createCmd.Flags().IntVarP(&Num, "number", "n", 0, "Number of VM for identification")
	createCmd.Flags().StringVarP(&image, "image", "i", "", "Image that VM will use (required)")
	createCmd.MarkFlagRequired("image")
	createCmd.Flags().IntVarP(&cpu, "cpu", "c", 2, "number of core")
	createCmd.Flags().IntVarP(&mem, "mem", "m", 4, "size of memory (GB)")
	createCmd.Flags().StringVarP(&userData, "user-data", "u", Datadir+"/cloudinit/user-data", "cloud-init user-data")
	createCmd.Flags().StringVarP(&metaData, "meta-data", "d", Datadir+"/cloudinit/meta-data", "cloud-init meta-data")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// /home/yjwang/문서/projects/golang/go-cli/src/virt-go/samples/focal-server-cloudimg-amd64.img
}
