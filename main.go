package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	if len(args) < 1 {
		fmt.Println("provide args!")
		return
	}

	paths := map[string]string{
		"module":     "src/infrastructure/modules",
		"controller": "src/infrastructure/http/controllers",
		"service":    "src/application/services",
		"irepo":      "src/domain/interfaces",
		"repository": "src/infrastructure/database/repositories",
		"entity":     "src/domain/entities",
		"model":      "src/infrastructure/database/models",
	}

	if args[1] == "gen" {
		gebbing := args[2:]
		fmt.Println("Generating files for the ff layers... ", gebbing)

		moduleName := input("Provide module name: ")
		moduleNameFormats := parseModuleName(moduleName)

		for _, layer := range args {
			fileName := filepath.Join(paths[layer], fmt.Sprintf("%s.%s.ts", moduleNameFormats["snake"], layer))
			switch layer {
			case "entity":
				createEntity(fileName, moduleNameFormats)
			case "model":
				createModel(fileName, moduleNameFormats)
			case "repository":
				createRepo(fileName, moduleNameFormats)
			case "irepo":
				irepoFileName := filepath.Join(paths[layer], fmt.Sprintf("I%sRepository.ts", moduleNameFormats["pascal"]))
				createIRepo(irepoFileName, moduleNameFormats)
			case "service":
				createService(fileName, moduleNameFormats)
			case "controller":
				createController(fileName, moduleNameFormats)
			case "module":
				createModule(fileName, moduleNameFormats)
			}
		}
	} else if args[1] == "spawn" {
		fmt.Println("spawning...")

		if len(args) == 3 {
			moduleName := args[2]
			moduleNameFormats := parseModuleName(moduleName)

			for layer, path := range paths {

				fileName := filepath.Join(path, fmt.Sprintf("%s.%s.ts", moduleNameFormats["snake"], layer))

				switch layer {
				case "entity":
					createEntity(fileName, moduleNameFormats)
				case "model":
					createModel(fileName, moduleNameFormats)
				case "irepo":
					fileName = filepath.Join(path, fmt.Sprintf("I%sRepository.ts", moduleNameFormats["pascal"]))
					createIRepo(fileName, moduleNameFormats)
				case "repository":
					createRepo(fileName, moduleNameFormats)
				case "service":
					createService(fileName, moduleNameFormats)
				case "controller":
					createController(fileName, moduleNameFormats)
				case "module":
					createModule(fileName, moduleNameFormats)
				}

			}
		} else {
			fmt.Println("provide module filename!")
			return
		}
	} else {
		fmt.Println("unknown command")
		return
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func parseModuleName(moduleName string) map[string]string {
	splitted := strings.Split(moduleName, "-")

	var camelCase string
	var pascalCase string
	for i := 0; i < len(splitted); i++ {
		if i > 0 {
			camelCase += capitalize(splitted[i])
		} else {
			camelCase += splitted[i]
		}
		pascalCase += capitalize(splitted[i])
	}

	formats := map[string]string{
		"pascal": pascalCase,
		"camel":  camelCase,
		"snake":  moduleName,
	}

	fmt.Println("Parsed module name in multiple formats: ", formats)

	return formats
}

func createEntity(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`export class %sEntity {

    id: string;
    
    private constructor(data: Partial<%sEntity>) { 
        Object.assign(this, data);
    }
        
    static create(data: Partial<%sEntity>): %sEntity {
        return new %sEntity(data);
    }
}`,
		mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"])

	write(fname, content)
}

func createModel(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { BaseModel } from './base.model';
import { PrimaryGeneratedColumn } from 'typeorm';

export class %s extends BaseModel {
    @PrimaryGeneratedColumn('uuid')
    id: string;
j}`, mdf["pascal"])

	write(fname, content)
}

func createIRepo(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { IBaseRepository } from './IBaseRepository';
import { %sEntity } from '../entities/%s.entity';
    
export abstract class I%sRepository extends IBaseRepository<%sEntity> {}`, mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["pascal"])
	write(fname, content)
}

func createRepo(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { I%sRepository } from '@/domain/interfaces/I%sRepository';
import { BaseRepository } from './base.repository';
import { %sEntity } from '@/domain/entities/%s.entity';
import { %s } from '@/infrastructure/database/models/%s.model';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
    
export class %sRepository extends BaseRepository<%sEntity, %s> implements I%sRepository {
	constructor(@InjectRepository(%s) repository: Repository<%s>) {
		super(
			{ entity: %sEntity, model: %s },
			repository,
			'' // CHANGE THIS
		)
	}
}`,
		mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"],
	)

	write(fname, content)
}

func createService(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { %sRepository } from '@/infrastructure/database/repositories/%s.repository';
import { Injectable, Inject } from '@nestjs/common';
import { I%sRepository } from '@/domain/interfaces/I%sRepository'

@Injectable()
export class %sService {
	constructor(@Inject(I%sRepository) private readonly %sRepository: I%sRepository) {}
}
    `, mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["camel"], mdf["pascal"])

	write(fname, content)
}

func createController(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { Controller } from '@nestjs/common';
import { %sService } from '@/application/services/%s.service'
import { Inject } from '@nestjs/common';

@Controller('endpot') // CHANGE THIS
export class %sController {
	constructor(@Inject(%sService) private readonly %sService: %sService) {}
}
`, mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["pascal"], mdf["camel"], mdf["pascal"])

	write(fname, content)
}

func createModule(fname string, mdf map[string]string) {
	content := fmt.Sprintf(`import { Module } from '@nestjs/common';
import { %sController } from '@/infrastructure/http/%s.controllers';
import { %sService } from '@/application/services/%s.service';
import { %sRepository } from '@/infrastructure/database/repositories/%s.repository';
import { I%sRepository } from '@/domain/interfaces/I%sRepository';

@Module({
	imports: [],
	providers: [
		{
			provide: I%sRepository,
			useClass: %sRepository
		},
		%sService
	],
	controllers: [%sController],
	exports: []
})
export class %sModule {}
`,
		mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["snake"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"], mdf["pascal"],
	)

	write(fname, content)
}

func write(fname string, content string) {
	file, err := os.Create(fname)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("created\t", fname)
}

func input(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
