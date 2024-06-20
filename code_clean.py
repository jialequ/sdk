import json
import re
import random
import string
import hashlib

str_set = string.digits

str_java_exception = "\n\tprivate class MultiPathRuntimeError extends RuntimeException{\n " + "\t\tpublic MultiPathRuntimeError(String error){\n\t\t\tsuper(error);\n\t\t}\n" + "\t\tpublic MultiPathRuntimeError(String error,Throwable e){\n\t\t\tsuper(error,e);\n\t\t}\n" +"\t}\n"

str_go_literal = "literal_"

global lines_dict

lines_dict = dict()

global func_nested

func_nested = 0

func_nested_list =[]

func_nested_file_list =[]

test_file_list =[]

#literal error
def solve1192(problem):
    # file_path = problem["resource"].replace("/","",1)
    file_path = problem["resource"]
    message = problem["message"]
    literal = getLiteral(message)
    print("clean " + file_path + " | " + "S1192: " + message)
    with open(file_path, mode="r", encoding="utf-8") as f:
        codes = f.read()

    if codes.find(literal)<0 or (codes.rfind("const "+str_go_literal) >0 and codes.rfind("const "+str_go_literal) < codes.rfind(literal)):
        return
    
    literal_def = str_go_literal + "".join(random.sample(str_set,4))
    
    #new_codes = re.sub(literal,changeLiteral,codes)

    new_codes = codes.replace("\""+literal+"\"",literal_def)

    new_codes += "\nconst " + literal_def + " = \"" + literal + "\"\n"
    
    with open(file_path, mode="w", encoding="utf-8") as f:
        f.write(new_codes)


#cognitive complexity error
def solve3776(problem):
    global lines_dict
    file_path = problem["resource"].replace("/","",1)
    file_key = hashlib.md5(file_path.encode("utf-8")).hexdigest()
    message = problem["message"]
    if file_path.find("test")>=0:
        test_file_list.append(file_path)
        return
    print("clean " + file_path + " | " + "S3776: " + message)
    new_codes = ""
    offset = 0
    if file_key not in lines_dict:
        lines_dict[file_key] = 0
    else:
        offset = lines_dict[file_key]
        
    with open(file_path, mode="r", encoding="utf-8") as f:
        lines = f.readlines()
        start, line_num, codes = findCodes(lines,offset+problem["startLineNumber"]-1)
        index = 0
        add_lines = 0
        while index < len(lines):
            if lines[index].find("package")>=0 and lines[index].find("test")>=0:
                test_file_list.append(file_path)
                return
            if(index == start):
                tmp_code, status, new_lines = reduceComplexity(lines,start, line_num, file_path)
                if status < 0:
                    return
                new_codes += tmp_code
                add_lines += new_lines
                index = start + line_num
            else :
                new_codes += lines[index]
                index = index + 1

    with open(file_path, mode="w", encoding="utf-8") as f:
        f.write(new_codes)

    lines_dict[file_key] += add_lines


#cognitive complexity error
def solve3776_2(problem):
    global lines_dict
    file_path = problem["resource"].replace("/","",1)
    file_key = hashlib.md5(file_path.encode("utf-8")).hexdigest()
    message = problem["message"]
    print("clean " + file_path + " | " + "S3776: " + message)
    new_codes = ""
    add_codes = ""
    offset = 0
    if file_key not in lines_dict:
        lines_dict[file_key] = 0
    else:
        offset = lines_dict[file_key]
        
    with open(file_path, mode="r", encoding="utf-8") as f:
        lines = f.readlines()
        start, line_num, codes = findCodes(lines,offset+problem["startLineNumber"]-1)
        index = 0
        add_lines = 0
        while index < len(lines):
            if(index == start):
                tmp_code, status, new_lines, add_code = reduceComplexity_func(lines,start, line_num, file_path)
                if status < 0:
                    return
                new_codes += tmp_code
                add_codes += add_code
                add_lines += new_lines
                index = start + line_num
            else :
                new_codes += lines[index]
                index = index + 1

    with open(file_path, mode="w", encoding="utf-8") as f:
        f.write(new_codes)
        f.write("\n")
        f.write(add_codes)

    lines_dict[file_key] += add_lines




#empty function error
def solve1186(problem):
    global lines_dict
    file_path = problem["resource"].replace("/","",1)
    file_key = hashlib.md5(file_path.encode("utf-8")).hexdigest()
    message = problem["message"]
    print("clean " + file_path + " | " + "S1186: " + message)
    new_codes = ""
    offset = 0
    if file_key not in lines_dict:
        lines_dict[file_key] = 0
    else:
        offset = lines_dict[file_key]
        
    with open(file_path, mode="r", encoding="utf-8") as f:
        lines = f.readlines()
        start, line_num, codes = findCodes(lines,offset+problem["startLineNumber"]-1)
        index = 0
        add_lines = 0
        while index < len(lines):
            if(index == start):
                tmp_code,new_lines = fillEmpty(codes,line_num)
                new_codes += tmp_code
                add_lines += new_lines
                index = start + line_num
            else :
                new_codes += lines[index]
                index = index + 1

    with open(file_path, mode="w", encoding="utf-8") as f:
        f.write(new_codes)

    lines_dict[file_key] += add_lines

#java exception error
def solve112(problem):
    file_path = problem["resource"].replace("/","",1)
    message = problem["message"]
    print("clean " + file_path + " | " + "S112: " + message)
    new_codes = ""
    with open(file_path, mode="r", encoding="utf-8") as f:
        codes = f.read()

    if codes.find("RuntimeException(") < 0:
        return

    new_codes = codes.replace("RuntimeException(","MultiPathRuntimeError(")

    end_pos = new_codes.rfind("}")

    new_codes = new_codes[:end_pos] + str_java_exception + new_codes[end_pos:-1]

    with open(file_path, mode="w", encoding="utf-8") as f:
        f.write(new_codes)


def reduceComplexity_func(lines,start, line_num, file_path):
    global func_nested
    reduce_level = 3
    index = start
    new_code = ""
    add_code = ""
    new_lines = 0;
    func_num = 0;
    while index < start + line_num:
        col = lines[index].find("func(")
        if index > start and col >= 0 and lines[index].find("{")>0 and lines[index].find("if") <0:
            new_func = "NestedFunc"+ "".join(random.sample(str_set,4))
            new_code += lines[index][:col]+ new_func + ")\n"
            new_lines += 1
            codes, func_lines = findEndFunc(lines,col, index,start+line_num, new_func)
            if func_lines < 0:
                print("###################error##################")
                print(lines[index])
                return new_code, -1, 0, ""
            add_code += codes
            new_lines -= func_lines
            index += func_lines
        else:
            new_code += lines[index]
            index += 1
    #print(new_code)
    #print(add_code)
    print(add_code)
    print(func_lines)
    func_nested += func_num 
    return new_code, 1, new_lines, add_code



def reduceComplexity(lines,start, line_num, file_path):
    global func_nested
    reduce_level = 3
    index = start
    new_code = ""
    new_lines = 0;
    func_num = 0;
    while index < start + line_num:
        col = lines[index].find("if ")
        if col >= reduce_level and lines[index].find("else ")<0 and lines[index].find("//")<0 and lines[index - 1].find("/*") < 0:
            new_code += "\t"*col + "/*\n"
            new_lines += 1
            codes, if_lines = findEndIf(lines,col, index,start+line_num)
            if if_lines < 0:
                print("###################error##################")
                print(lines[index])
                return new_code, -1, 0
            new_code += codes.replace("/*","").replace("*/","") +  "\t"*col + "*/\n"
            new_lines += 1
            index += if_lines
        else:
            new_code += lines[index]
            index += 1
    #print(new_code)
    func_nested += func_num 
    return new_code, 1, new_lines


def findEndIf(lines,col, index, end):
    codes = ""
    if_lines = 1
    while index < end:
        if len(lines[index])>col and lines[index][col]=="}" and lines[index].find("else") < 0 and lines[index].find("//")<0:
            codes += lines[index]
            return codes, if_lines
        else:
            codes += lines[index]
            if_lines +=1
            index += 1
    print("###################error##################")
    print(codes)
    return codes, -1
        
def findEndFunc(lines,col, start, end, func_name):
    codes = ""
    index = start+1
    func_lines = 2
    pos = lines[start].rfind("\t")+1
    codes += lines[start][col:-1].replace("func(", "func "+func_name+"(") + "\n"
    while index < end:
        if len(lines[index])>pos and lines[index][pos]=="}" and lines[index][pos+1] == ")":
            codes += "}\n\n"
            return codes, func_lines
        else:
            codes += lines[index].replace("\t"*pos,"",1)
            func_lines +=1
            index += 1
    #print("###################error##################")
    #print(codes)
    return codes, -1

# fill the case 1 line func or 2 line func
def fillEmpty(codes,line_num):
    new_lines = 0
    if line_num == 1:
        ref = "  //"+codes.replace("{","").replace("}","").replace("\n","")
        codes = codes.replace("}","\n" + ref + "\n" + "}")
        new_lines += 2
    if line_num == 2:
        ref = "  //"+codes.replace("{","").replace("}","").replace("\n","")
        codes = codes.replace("}","\n" + ref + "\n" + "}")
        new_lines += 2
    return codes, new_lines

def findCodes(lines,start):
    while start>=0:
        if lines[start].find("func ")>=0:
            break
        else:
           start = start - 1
    codes = lines[start]
    if codes.find("}")>0:
        return start, 1, codes
    pos = start + 1
    while pos<len(lines):
        codes += lines[pos]
        if lines[pos][0]=="}":
            break
        else:
            pos = pos + 1
    return start, pos - start + 1, codes
    


def changeLiteral(matched):
    output = "".join(random.sample(str_set,4)) + matched.group() + " "+ "".join(random.sample(str_set,2))
    #print(output)
    return output
        


def getLiteral(message):
    pos1 = message.find("this literal \"")
    pos2 = message.rfind("\"")
    return message[pos1+14:pos2]



def main():
    print("-------begin code clean!-------")
    f = open("problems.txt",encoding = "utf-8")
    problems = json.load(f)
    f.close()

    for problem in problems:
       code = problem["code"]
       if code == "go:S1192":
          solve1192(problem)
          continue
    #    if code == "go:S3776":
    #       solve3776_2(problem)
       #if code == "go:S1186":
          #solve1186(problem)
       #if code == "java:S112":
          #solve112(problem)
          
    print("-------end code clean!---------")
    print("func_nested num is " + str(func_nested))
    print(func_nested_list)
    print("test func num is " + str(len(func_nested_file_list)))
    test_num = 0
    non_test_num = 0
    for file_name in set(test_file_list):
        print(file_name)
        if file_name.find("test")>=0:
            test_num += 1
        else:
            non_test_num +=1
    print(test_num)
    print(non_test_num)
     
if __name__== "__main__" :
    main()    
  


